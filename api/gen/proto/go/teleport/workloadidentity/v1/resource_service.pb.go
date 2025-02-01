// Copyright 2024 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        (unknown)
// source: teleport/workloadidentity/v1/resource_service.proto

package workloadidentityv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// The request for CreateWorkloadIdentity.
type CreateWorkloadIdentityRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The workload identity to create.
	WorkloadIdentity *WorkloadIdentity `protobuf:"bytes,1,opt,name=workload_identity,json=workloadIdentity,proto3" json:"workload_identity,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *CreateWorkloadIdentityRequest) Reset() {
	*x = CreateWorkloadIdentityRequest{}
	mi := &file_teleport_workloadidentity_v1_resource_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateWorkloadIdentityRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateWorkloadIdentityRequest) ProtoMessage() {}

func (x *CreateWorkloadIdentityRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_workloadidentity_v1_resource_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateWorkloadIdentityRequest.ProtoReflect.Descriptor instead.
func (*CreateWorkloadIdentityRequest) Descriptor() ([]byte, []int) {
	return file_teleport_workloadidentity_v1_resource_service_proto_rawDescGZIP(), []int{0}
}

func (x *CreateWorkloadIdentityRequest) GetWorkloadIdentity() *WorkloadIdentity {
	if x != nil {
		return x.WorkloadIdentity
	}
	return nil
}

// The request for UpdateWorkloadIdentity.
type UpdateWorkloadIdentityRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The workload identity to update.
	WorkloadIdentity *WorkloadIdentity `protobuf:"bytes,1,opt,name=workload_identity,json=workloadIdentity,proto3" json:"workload_identity,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *UpdateWorkloadIdentityRequest) Reset() {
	*x = UpdateWorkloadIdentityRequest{}
	mi := &file_teleport_workloadidentity_v1_resource_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateWorkloadIdentityRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateWorkloadIdentityRequest) ProtoMessage() {}

func (x *UpdateWorkloadIdentityRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_workloadidentity_v1_resource_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateWorkloadIdentityRequest.ProtoReflect.Descriptor instead.
func (*UpdateWorkloadIdentityRequest) Descriptor() ([]byte, []int) {
	return file_teleport_workloadidentity_v1_resource_service_proto_rawDescGZIP(), []int{1}
}

func (x *UpdateWorkloadIdentityRequest) GetWorkloadIdentity() *WorkloadIdentity {
	if x != nil {
		return x.WorkloadIdentity
	}
	return nil
}

// The request for UpsertWorkloadIdentityRequest.
type UpsertWorkloadIdentityRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The workload identity to upsert.
	WorkloadIdentity *WorkloadIdentity `protobuf:"bytes,1,opt,name=workload_identity,json=workloadIdentity,proto3" json:"workload_identity,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *UpsertWorkloadIdentityRequest) Reset() {
	*x = UpsertWorkloadIdentityRequest{}
	mi := &file_teleport_workloadidentity_v1_resource_service_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpsertWorkloadIdentityRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertWorkloadIdentityRequest) ProtoMessage() {}

func (x *UpsertWorkloadIdentityRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_workloadidentity_v1_resource_service_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertWorkloadIdentityRequest.ProtoReflect.Descriptor instead.
func (*UpsertWorkloadIdentityRequest) Descriptor() ([]byte, []int) {
	return file_teleport_workloadidentity_v1_resource_service_proto_rawDescGZIP(), []int{2}
}

func (x *UpsertWorkloadIdentityRequest) GetWorkloadIdentity() *WorkloadIdentity {
	if x != nil {
		return x.WorkloadIdentity
	}
	return nil
}

// The request for GetWorkloadIdentity.
type GetWorkloadIdentityRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The name of the workload identity to retrieve.
	Name          string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetWorkloadIdentityRequest) Reset() {
	*x = GetWorkloadIdentityRequest{}
	mi := &file_teleport_workloadidentity_v1_resource_service_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetWorkloadIdentityRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetWorkloadIdentityRequest) ProtoMessage() {}

func (x *GetWorkloadIdentityRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_workloadidentity_v1_resource_service_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetWorkloadIdentityRequest.ProtoReflect.Descriptor instead.
func (*GetWorkloadIdentityRequest) Descriptor() ([]byte, []int) {
	return file_teleport_workloadidentity_v1_resource_service_proto_rawDescGZIP(), []int{3}
}

func (x *GetWorkloadIdentityRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

// The request for DeleteWorkloadIdentity.
type DeleteWorkloadIdentityRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The name of the workload identity to delete.
	Name          string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteWorkloadIdentityRequest) Reset() {
	*x = DeleteWorkloadIdentityRequest{}
	mi := &file_teleport_workloadidentity_v1_resource_service_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteWorkloadIdentityRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteWorkloadIdentityRequest) ProtoMessage() {}

func (x *DeleteWorkloadIdentityRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_workloadidentity_v1_resource_service_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteWorkloadIdentityRequest.ProtoReflect.Descriptor instead.
func (*DeleteWorkloadIdentityRequest) Descriptor() ([]byte, []int) {
	return file_teleport_workloadidentity_v1_resource_service_proto_rawDescGZIP(), []int{4}
}

func (x *DeleteWorkloadIdentityRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

// The request for ListWorkloadIdentities.
type ListWorkloadIdentitiesRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The maximum number of items to return.
	// The server may impose a different page size at its discretion.
	PageSize int32 `protobuf:"varint,1,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	// The page_token value returned from a previous ListWorkloadIdentities request, if any.
	PageToken     string `protobuf:"bytes,2,opt,name=page_token,json=pageToken,proto3" json:"page_token,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListWorkloadIdentitiesRequest) Reset() {
	*x = ListWorkloadIdentitiesRequest{}
	mi := &file_teleport_workloadidentity_v1_resource_service_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListWorkloadIdentitiesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListWorkloadIdentitiesRequest) ProtoMessage() {}

func (x *ListWorkloadIdentitiesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_workloadidentity_v1_resource_service_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListWorkloadIdentitiesRequest.ProtoReflect.Descriptor instead.
func (*ListWorkloadIdentitiesRequest) Descriptor() ([]byte, []int) {
	return file_teleport_workloadidentity_v1_resource_service_proto_rawDescGZIP(), []int{5}
}

func (x *ListWorkloadIdentitiesRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

func (x *ListWorkloadIdentitiesRequest) GetPageToken() string {
	if x != nil {
		return x.PageToken
	}
	return ""
}

// The response for ListWorkloadIdentities.
type ListWorkloadIdentitiesResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The page of workload identities that matched the request.
	WorkloadIdentities []*WorkloadIdentity `protobuf:"bytes,1,rep,name=workload_identities,json=workloadIdentities,proto3" json:"workload_identities,omitempty"`
	// Token to retrieve the next page of results, or empty if there are no
	// more results in the list.
	NextPageToken string `protobuf:"bytes,2,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListWorkloadIdentitiesResponse) Reset() {
	*x = ListWorkloadIdentitiesResponse{}
	mi := &file_teleport_workloadidentity_v1_resource_service_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListWorkloadIdentitiesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListWorkloadIdentitiesResponse) ProtoMessage() {}

func (x *ListWorkloadIdentitiesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_workloadidentity_v1_resource_service_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListWorkloadIdentitiesResponse.ProtoReflect.Descriptor instead.
func (*ListWorkloadIdentitiesResponse) Descriptor() ([]byte, []int) {
	return file_teleport_workloadidentity_v1_resource_service_proto_rawDescGZIP(), []int{6}
}

func (x *ListWorkloadIdentitiesResponse) GetWorkloadIdentities() []*WorkloadIdentity {
	if x != nil {
		return x.WorkloadIdentities
	}
	return nil
}

func (x *ListWorkloadIdentitiesResponse) GetNextPageToken() string {
	if x != nil {
		return x.NextPageToken
	}
	return ""
}

var File_teleport_workloadidentity_v1_resource_service_proto protoreflect.FileDescriptor

var file_teleport_workloadidentity_v1_resource_service_proto_rawDesc = string([]byte{
	0x0a, 0x33, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x6c,
	0x6f, 0x61, 0x64, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2f, 0x76, 0x31, 0x2f, 0x72,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1c, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e,
	0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x2b, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x6c,
	0x6f, 0x61, 0x64, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2f, 0x76, 0x31, 0x2f, 0x72,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x7c, 0x0a,
	0x1d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49,
	0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x5b,
	0x0a, 0x11, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x69, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x74, 0x65, 0x6c, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x69, 0x64, 0x65,
	0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61,
	0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x10, 0x77, 0x6f, 0x72, 0x6b, 0x6c,
	0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x22, 0x7c, 0x0a, 0x1d, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65,
	0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x5b, 0x0a, 0x11,
	0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f,
	0x72, 0x74, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x69, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49,
	0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x10, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61,
	0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x22, 0x7c, 0x0a, 0x1d, 0x55, 0x70, 0x73,
	0x65, 0x72, 0x74, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x5b, 0x0a, 0x11, 0x77, 0x6f,
	0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74,
	0x79, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65,
	0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x10, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49,
	0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x22, 0x30, 0x0a, 0x1a, 0x47, 0x65, 0x74, 0x57, 0x6f,
	0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x33, 0x0a, 0x1d, 0x44, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x5b,
	0x0a, 0x1d, 0x4c, 0x69, 0x73, 0x74, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64,
	0x65, 0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x1b, 0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x1d, 0x0a, 0x0a,
	0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x70, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0xa9, 0x01, 0x0a, 0x1e,
	0x4c, 0x69, 0x73, 0x74, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x5f,
	0x0a, 0x13, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x69, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x69, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x74, 0x65,
	0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x69,
	0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x6c,
	0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x12, 0x77, 0x6f, 0x72,
	0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x12,
	0x26, 0x0a, 0x0f, 0x6e, 0x65, 0x78, 0x74, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b,
	0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x6e, 0x65, 0x78, 0x74, 0x50, 0x61,
	0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x32, 0xbf, 0x06, 0x0a, 0x1f, 0x57, 0x6f, 0x72, 0x6b,
	0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x85, 0x01, 0x0a, 0x16,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64,
	0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x3b, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72,
	0x74, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69,
	0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b,
	0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x2e, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x77,
	0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e,
	0x76, 0x31, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x12, 0x85, 0x01, 0x0a, 0x16, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x57, 0x6f,
	0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x3b,
	0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f,
	0x61, 0x64, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2e, 0x2e, 0x74, 0x65,
	0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x69,
	0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x6c,
	0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x85, 0x01, 0x0a, 0x16,
	0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64,
	0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x3b, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72,
	0x74, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69,
	0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x57, 0x6f, 0x72, 0x6b,
	0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x2e, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x77,
	0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e,
	0x76, 0x31, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x12, 0x7f, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f,
	0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x38, 0x2e, 0x74, 0x65, 0x6c,
	0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x69, 0x64,
	0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x57, 0x6f, 0x72,
	0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x2e, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e,
	0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x2e, 0x76, 0x31, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x12, 0x6d, 0x0a, 0x16, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x57, 0x6f,
	0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x3b,
	0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f,
	0x61, 0x64, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x12, 0x93, 0x01, 0x0a, 0x16, 0x4c, 0x69, 0x73, 0x74, 0x57, 0x6f, 0x72, 0x6b,
	0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x12, 0x3b,
	0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f,
	0x61, 0x64, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69,
	0x73, 0x74, 0x57, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69,
	0x74, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x3c, 0x2e, 0x74, 0x65,
	0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x69,
	0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x57,
	0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x69, 0x65,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x64, 0x5a, 0x62, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x72, 0x61, 0x76, 0x69, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f,
	0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61,
	0x64, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2f, 0x76, 0x31, 0x3b, 0x77, 0x6f, 0x72,
	0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x76, 0x31, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_teleport_workloadidentity_v1_resource_service_proto_rawDescOnce sync.Once
	file_teleport_workloadidentity_v1_resource_service_proto_rawDescData []byte
)

func file_teleport_workloadidentity_v1_resource_service_proto_rawDescGZIP() []byte {
	file_teleport_workloadidentity_v1_resource_service_proto_rawDescOnce.Do(func() {
		file_teleport_workloadidentity_v1_resource_service_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_teleport_workloadidentity_v1_resource_service_proto_rawDesc), len(file_teleport_workloadidentity_v1_resource_service_proto_rawDesc)))
	})
	return file_teleport_workloadidentity_v1_resource_service_proto_rawDescData
}

var file_teleport_workloadidentity_v1_resource_service_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_teleport_workloadidentity_v1_resource_service_proto_goTypes = []any{
	(*CreateWorkloadIdentityRequest)(nil),  // 0: teleport.workloadidentity.v1.CreateWorkloadIdentityRequest
	(*UpdateWorkloadIdentityRequest)(nil),  // 1: teleport.workloadidentity.v1.UpdateWorkloadIdentityRequest
	(*UpsertWorkloadIdentityRequest)(nil),  // 2: teleport.workloadidentity.v1.UpsertWorkloadIdentityRequest
	(*GetWorkloadIdentityRequest)(nil),     // 3: teleport.workloadidentity.v1.GetWorkloadIdentityRequest
	(*DeleteWorkloadIdentityRequest)(nil),  // 4: teleport.workloadidentity.v1.DeleteWorkloadIdentityRequest
	(*ListWorkloadIdentitiesRequest)(nil),  // 5: teleport.workloadidentity.v1.ListWorkloadIdentitiesRequest
	(*ListWorkloadIdentitiesResponse)(nil), // 6: teleport.workloadidentity.v1.ListWorkloadIdentitiesResponse
	(*WorkloadIdentity)(nil),               // 7: teleport.workloadidentity.v1.WorkloadIdentity
	(*emptypb.Empty)(nil),                  // 8: google.protobuf.Empty
}
var file_teleport_workloadidentity_v1_resource_service_proto_depIdxs = []int32{
	7,  // 0: teleport.workloadidentity.v1.CreateWorkloadIdentityRequest.workload_identity:type_name -> teleport.workloadidentity.v1.WorkloadIdentity
	7,  // 1: teleport.workloadidentity.v1.UpdateWorkloadIdentityRequest.workload_identity:type_name -> teleport.workloadidentity.v1.WorkloadIdentity
	7,  // 2: teleport.workloadidentity.v1.UpsertWorkloadIdentityRequest.workload_identity:type_name -> teleport.workloadidentity.v1.WorkloadIdentity
	7,  // 3: teleport.workloadidentity.v1.ListWorkloadIdentitiesResponse.workload_identities:type_name -> teleport.workloadidentity.v1.WorkloadIdentity
	0,  // 4: teleport.workloadidentity.v1.WorkloadIdentityResourceService.CreateWorkloadIdentity:input_type -> teleport.workloadidentity.v1.CreateWorkloadIdentityRequest
	1,  // 5: teleport.workloadidentity.v1.WorkloadIdentityResourceService.UpdateWorkloadIdentity:input_type -> teleport.workloadidentity.v1.UpdateWorkloadIdentityRequest
	2,  // 6: teleport.workloadidentity.v1.WorkloadIdentityResourceService.UpsertWorkloadIdentity:input_type -> teleport.workloadidentity.v1.UpsertWorkloadIdentityRequest
	3,  // 7: teleport.workloadidentity.v1.WorkloadIdentityResourceService.GetWorkloadIdentity:input_type -> teleport.workloadidentity.v1.GetWorkloadIdentityRequest
	4,  // 8: teleport.workloadidentity.v1.WorkloadIdentityResourceService.DeleteWorkloadIdentity:input_type -> teleport.workloadidentity.v1.DeleteWorkloadIdentityRequest
	5,  // 9: teleport.workloadidentity.v1.WorkloadIdentityResourceService.ListWorkloadIdentities:input_type -> teleport.workloadidentity.v1.ListWorkloadIdentitiesRequest
	7,  // 10: teleport.workloadidentity.v1.WorkloadIdentityResourceService.CreateWorkloadIdentity:output_type -> teleport.workloadidentity.v1.WorkloadIdentity
	7,  // 11: teleport.workloadidentity.v1.WorkloadIdentityResourceService.UpdateWorkloadIdentity:output_type -> teleport.workloadidentity.v1.WorkloadIdentity
	7,  // 12: teleport.workloadidentity.v1.WorkloadIdentityResourceService.UpsertWorkloadIdentity:output_type -> teleport.workloadidentity.v1.WorkloadIdentity
	7,  // 13: teleport.workloadidentity.v1.WorkloadIdentityResourceService.GetWorkloadIdentity:output_type -> teleport.workloadidentity.v1.WorkloadIdentity
	8,  // 14: teleport.workloadidentity.v1.WorkloadIdentityResourceService.DeleteWorkloadIdentity:output_type -> google.protobuf.Empty
	6,  // 15: teleport.workloadidentity.v1.WorkloadIdentityResourceService.ListWorkloadIdentities:output_type -> teleport.workloadidentity.v1.ListWorkloadIdentitiesResponse
	10, // [10:16] is the sub-list for method output_type
	4,  // [4:10] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_teleport_workloadidentity_v1_resource_service_proto_init() }
func file_teleport_workloadidentity_v1_resource_service_proto_init() {
	if File_teleport_workloadidentity_v1_resource_service_proto != nil {
		return
	}
	file_teleport_workloadidentity_v1_resource_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_teleport_workloadidentity_v1_resource_service_proto_rawDesc), len(file_teleport_workloadidentity_v1_resource_service_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_teleport_workloadidentity_v1_resource_service_proto_goTypes,
		DependencyIndexes: file_teleport_workloadidentity_v1_resource_service_proto_depIdxs,
		MessageInfos:      file_teleport_workloadidentity_v1_resource_service_proto_msgTypes,
	}.Build()
	File_teleport_workloadidentity_v1_resource_service_proto = out.File
	file_teleport_workloadidentity_v1_resource_service_proto_goTypes = nil
	file_teleport_workloadidentity_v1_resource_service_proto_depIdxs = nil
}
