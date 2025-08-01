/*
 * Teleport
 * Copyright (C) 2024  Gravitational, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package aws_sync

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/gravitational/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	accessgraphv1alpha "github.com/gravitational/teleport/gen/proto/go/accessgraph/v1alpha"
	libcloudaws "github.com/gravitational/teleport/lib/cloud/aws"
)

// pollAWSEC2Instances is a function that returns a function that fetches
// ec2 instances and instance profiles and returns an error if any.
func (a *Fetcher) pollAWSEC2Instances(ctx context.Context, result *Resources, collectErr func(error)) func() error {
	return func() error {
		var err error

		result.Instances, err = a.fetchAWSEC2Instances(ctx)
		if err != nil {
			collectErr(trace.Wrap(err, "failed to fetch instances"))
		}
		result.InstanceProfiles, err = a.fetchInstanceProfiles(ctx)
		if err != nil {
			collectErr(trace.Wrap(err, "failed to fetch instance profiles"))
		}
		return nil
	}
}

// fetchAWSEC2Instances fetches ec2 instances from all regions and returns them
// as a slice of accessgraphv1alpha.AWSInstanceV1.
// It uses ec2.DescribeInstancesPagesWithContext to iterate over all instances
// in all regions.
func (a *Fetcher) fetchAWSEC2Instances(ctx context.Context) ([]*accessgraphv1alpha.AWSInstanceV1, error) {
	var (
		hosts    []*accessgraphv1alpha.AWSInstanceV1
		hostsMu  sync.Mutex
		errs     []error
		existing = a.lastResult.Instances
	)
	eG, ctx := errgroup.WithContext(ctx)
	// Set the limit to 5 to avoid too many concurrent requests.
	// This is a temporary solution until we have a better way to limit the
	// number of concurrent requests.
	eG.SetLimit(5)
	collectHosts := func(lHosts []*accessgraphv1alpha.AWSInstanceV1, err error) {
		hostsMu.Lock()
		defer hostsMu.Unlock()
		if err != nil {
			errs = append(errs, err)
		}
		hosts = append(hosts, lHosts...)
	}

	for _, region := range a.Regions {
		eG.Go(func() error {
			prevIterationEc2 := sliceFilter(
				existing,
				func(h *accessgraphv1alpha.AWSInstanceV1) bool {
					return h.Region == region && h.AccountId == a.AccountID
				},
			)
			ec2Client, err := a.GetEC2Client(ctx, region, a.getAWSOptions()...)
			if err != nil {
				collectHosts(prevIterationEc2, trace.Wrap(err))
				return nil
			}
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			paginator := ec2.NewDescribeInstancesPaginator(ec2Client, &ec2.DescribeInstancesInput{
				MaxResults: aws.Int32(pageSize),
			})
			for paginator.HasMorePages() {
				page, err := paginator.NextPage(ctx)
				if err != nil {
					return libcloudaws.ConvertRequestFailureError(err)
				}
				lHosts := make([]*accessgraphv1alpha.AWSInstanceV1, 0, len(page.Reservations))
				for _, reservation := range page.Reservations {
					for _, instance := range reservation.Instances {
						hosts = append(hosts, awsInstanceToProtoInstance(instance, region, a.AccountID))
					}
				}
				collectHosts(lHosts, nil)
			}
			return nil
		})
	}

	err := eG.Wait()
	return hosts, trace.NewAggregate(append(errs, err)...)
}

// awsInstanceToProtoInstance converts an ec2.Instance to accessgraphv1alpha.AWSInstanceV1
// representation.
func awsInstanceToProtoInstance(instance ec2types.Instance, region string, accountID string) *accessgraphv1alpha.AWSInstanceV1 {
	var tags []*accessgraphv1alpha.AWSTag
	for _, tag := range instance.Tags {
		tags = append(tags, &accessgraphv1alpha.AWSTag{
			Key:   aws.ToString(tag.Key),
			Value: strPtrToWrapper(tag.Value),
		})
	}

	var instanceProfileMetadata *wrapperspb.StringValue
	if instance.IamInstanceProfile != nil {
		instanceProfileMetadata = strPtrToWrapper(instance.IamInstanceProfile.Arn)
	}
	return &accessgraphv1alpha.AWSInstanceV1{
		InstanceId:            aws.ToString(instance.InstanceId),
		Region:                region,
		PublicDnsName:         aws.ToString(instance.PublicDnsName),
		LaunchKeyName:         strPtrToWrapper(instance.KeyName),
		IamInstanceProfileArn: instanceProfileMetadata,
		AccountId:             accountID,
		Tags:                  tags,
		LaunchTime:            awsTimeToProtoTime(instance.LaunchTime),
		LastSyncTime:          timestamppb.Now(),
	}
}

// fetchInstanceProfiles fetches instance profiles from all regions and returns them
// as a slice of accessgraphv1alpha.AWSInstanceProfileV1.
func (a *Fetcher) fetchInstanceProfiles(ctx context.Context) ([]*accessgraphv1alpha.AWSInstanceProfileV1, error) {
	existing := a.lastResult.InstanceProfiles
	awsCfg, err := a.AWSConfigProvider.GetConfig(
		ctx,
		"", /* region is empty because users and groups are global */
		a.getAWSOptions()...,
	)
	if err != nil {
		return existing, trace.Wrap(err)
	}
	iamClient := a.awsClients.getIAMClient(awsCfg)
	pager := iam.NewListInstanceProfilesPaginator(
		iamClient,
		&iam.ListInstanceProfilesInput{
			MaxItems: aws.Int32(pageSize),
		},
		func(opts *iam.ListInstanceProfilesPaginatorOptions) {
			opts.StopOnDuplicateToken = true
		},
	)

	var profiles []*accessgraphv1alpha.AWSInstanceProfileV1
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return append(profiles, existing...), trace.Wrap(err)
		}
		for _, profile := range page.InstanceProfiles {
			profiles = append(profiles,
				awsInstanceProfileToProtoInstanceProfile(profile, a.AccountID),
			)
		}
	}
	return profiles, trace.Wrap(err)
}

// awsInstanceProfileToProtoInstanceProfile converts an iam.InstanceProfile to accessgraphv1alpha.AWSInstanceProfileV1
func awsInstanceProfileToProtoInstanceProfile(profile iamtypes.InstanceProfile, accountID string) *accessgraphv1alpha.AWSInstanceProfileV1 {
	tags := make([]*accessgraphv1alpha.AWSTag, 0, len(profile.Tags))
	for _, tag := range profile.Tags {
		tags = append(tags, &accessgraphv1alpha.AWSTag{
			Key:   aws.ToString(tag.Key),
			Value: strPtrToWrapper(tag.Value),
		})
	}

	out := &accessgraphv1alpha.AWSInstanceProfileV1{
		InstanceProfileId:   aws.ToString(profile.InstanceProfileId),
		InstanceProfileName: aws.ToString(profile.InstanceProfileName),
		Arn:                 aws.ToString(profile.Arn),
		Path:                aws.ToString(profile.Path),
		AccountId:           accountID,
		Tags:                tags,
		CreatedAt:           awsTimeToProtoTime(profile.CreateDate),
		LastSyncTime:        timestamppb.Now(),
	}
	for _, role := range profile.Roles {
		out.Roles = append(out.Roles, awsRoleToProtoRole(role, accountID))
	}
	return out
}

func sliceFilter[T any](s []T, f func(T) bool) []T {
	var out []T
	for _, v := range s {
		if f(v) {
			out = append(out, v)
		}
	}
	return out
}

func sliceFilterPickFirst[T any](s []T, f func(T) bool) T {
	for _, v := range s {
		if f(v) {
			return v
		}
	}
	var v T
	return v
}
