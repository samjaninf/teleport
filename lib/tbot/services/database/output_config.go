/*
 * Teleport
 * Copyright (C) 2025  Gravitational, Inc.
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

package database

import (
	"context"
	"slices"

	"github.com/gravitational/trace"
	"gopkg.in/yaml.v3"

	"github.com/gravitational/teleport/lib/tbot/bot"
	"github.com/gravitational/teleport/lib/tbot/bot/destination"
	"github.com/gravitational/teleport/lib/tbot/internal"
	"github.com/gravitational/teleport/lib/tbot/internal/encoding"
)

const OutputServiceType = "database"

// DatabaseFormat specifies if any special behavior should be invoked when
// producing artifacts. This allows for databases/clients that require unique
// formats or paths to be used.
type DatabaseFormat string

const (
	// UnspecifiedDatabaseFormat is the unset value and the default. This
	// should work for most databases.
	UnspecifiedDatabaseFormat DatabaseFormat = ""
	// TLSDatabaseFormat is for databases that require specifically named
	// outputs: tls.key, tls.crt and tls.cas
	TLSDatabaseFormat DatabaseFormat = "tls"
	// MongoDatabaseFormat indicates credentials should be generated which
	// are compatible with MongoDB.
	// This outputs `mongo.crt` and `mongo.cas`.
	MongoDatabaseFormat DatabaseFormat = "mongo"
	// CockroachDatabaseFormat indicates credentials should be generated which
	// are compatible with CockroachDB.
	// This outputs `cockroach/node.key`, `cockroach/node.crt` and
	// `cockroach/ca.crt`.
	CockroachDatabaseFormat DatabaseFormat = "cockroach"
)

const (
	// DefaultMongoPrefix is the default prefix in generated MongoDB certs.
	DefaultMongoPrefix      = "mongo"
	DefaultCockroachDirName = "cockroach"
)

var databaseFormats = []DatabaseFormat{
	UnspecifiedDatabaseFormat,
	TLSDatabaseFormat,
	MongoDatabaseFormat,
	CockroachDatabaseFormat,
}

// OutputConfig produces credentials which can be used to connect to a database
// through teleport.
type OutputConfig struct {
	// Name of the service for logs and the /readyz endpoint.
	Name string `yaml:"name,omitempty"`
	// Destination is where the credentials should be written to.
	Destination destination.Destination `yaml:"destination"`
	// Roles is the list of roles to request for the generated credentials.
	// If empty, it defaults to all the bot's roles.
	Roles []string `yaml:"roles,omitempty"`

	// Formats specifies if any special behavior should be invoked when
	// producing artifacts. An empty value is supported by most database,
	// but CockroachDB and MongoDB require this value to be set to
	// `mongo` and `cockroach` respectively.
	Format DatabaseFormat `yaml:"format,omitempty"`
	// Service is the service name of the Teleport database. Generally this is
	// the name of the Teleport resource. This field is required for all types
	// of database.
	Service string `yaml:"service"`
	// Database is the name of the database to request access to.
	Database string `yaml:"database,omitempty"`
	// Username is the database username to request access as.
	Username string `yaml:"username,omitempty"`

	// CredentialLifetime contains configuration for how long credentials will
	// last and the frequency at which they'll be renewed.
	CredentialLifetime bot.CredentialLifetime `yaml:",inline"`
}

// GetName returns the user-given name of the service, used for validation purposes.
func (o *OutputConfig) GetName() string {
	return o.Name
}

func (o *OutputConfig) Init(ctx context.Context) error {
	subDirs := []string{}
	if o.Format == CockroachDatabaseFormat {
		subDirs = append(subDirs, DefaultCockroachDirName)
	}
	return trace.Wrap(o.Destination.Init(ctx, subDirs))
}

func (o *OutputConfig) CheckAndSetDefaults() error {
	if o.Destination == nil {
		return trace.BadParameter("no destination configured for output")
	}
	if err := o.Destination.CheckAndSetDefaults(); err != nil {
		return trace.Wrap(err, "validating destination")
	}

	if o.Service == "" {
		return trace.BadParameter("service must not be empty")
	}

	if !slices.Contains(databaseFormats, o.Format) {
		return trace.BadParameter("unrecognized format (%s)", o.Format)
	}

	return nil
}

func (o *OutputConfig) GetDestination() destination.Destination {
	return o.Destination
}

func (o *OutputConfig) Describe() []bot.FileDescription {
	fds := []bot.FileDescription{
		{
			Name: internal.IdentityFilePath,
		},
		{
			Name: internal.HostCAPath,
		},
		{
			Name: internal.UserCAPath,
		},
		{
			Name: internal.DatabaseCAPath,
		},
	}
	switch o.Format {
	case MongoDatabaseFormat:
		fds = append(fds, []bot.FileDescription{
			{
				Name: DefaultMongoPrefix + ".crt",
			},
			{
				Name: DefaultMongoPrefix + ".cas",
			},
		}...)
	case CockroachDatabaseFormat:
		fds = append(fds, []bot.FileDescription{
			{
				Name:  DefaultCockroachDirName,
				IsDir: true,
			},
		}...)
	case TLSDatabaseFormat:
		fds = append(fds, []bot.FileDescription{
			{
				Name: internal.DefaultTLSPrefix + ".crt",
			},
			{
				Name: internal.DefaultTLSPrefix + ".key",
			},
			{
				Name: internal.DefaultTLSPrefix + ".cas",
			},
		}...)
	}

	return fds
}

func (o *OutputConfig) MarshalYAML() (any, error) {
	type raw OutputConfig
	return encoding.WithTypeHeader((*raw)(o), OutputServiceType)
}

func (o *OutputConfig) UnmarshalYAML(*yaml.Node) error {
	return trace.NotImplemented("unmarshaling %T with UnmarshalYAML is not supported, use UnmarshalConfig instead", o)
}

func (o *OutputConfig) UnmarshalConfig(ctx bot.UnmarshalConfigContext, node *yaml.Node) error {
	dest, err := internal.ExtractOutputDestination(ctx, node)
	if err != nil {
		return trace.Wrap(err)
	}
	// Alias type to remove UnmarshalYAML to avoid getting our "not implemented" error
	type raw OutputConfig
	if err := node.Decode((*raw)(o)); err != nil {
		return trace.Wrap(err)
	}
	o.Destination = dest
	return nil
}

func (o *OutputConfig) Type() string {
	return OutputServiceType
}

func (o *OutputConfig) GetCredentialLifetime() bot.CredentialLifetime {
	return o.CredentialLifetime
}

// SupportedDatabaseFormatStrings returns a constant list of all valid
// DatabaseFormat values as strings.
func SupportedDatabaseFormatStrings() (ret []string) {
	for _, v := range databaseFormats {
		ret = append(ret, string(v))
	}

	return
}
