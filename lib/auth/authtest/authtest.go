/*
 * Teleport
 * Copyright (C) 2023  Gravitational, Inc.
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

package authtest

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/google/uuid"
	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/api/breaker"
	"github.com/gravitational/teleport/api/client"
	"github.com/gravitational/teleport/api/client/proto"
	"github.com/gravitational/teleport/api/constants"
	apidefaults "github.com/gravitational/teleport/api/defaults"
	"github.com/gravitational/teleport/api/types"
	apiutils "github.com/gravitational/teleport/api/utils"
	"github.com/gravitational/teleport/api/utils/keys"
	"github.com/gravitational/teleport/lib/auth"
	"github.com/gravitational/teleport/lib/auth/accesspoint"
	"github.com/gravitational/teleport/lib/auth/authclient"
	"github.com/gravitational/teleport/lib/auth/state"
	authority "github.com/gravitational/teleport/lib/auth/testauthority"
	"github.com/gravitational/teleport/lib/authz"
	"github.com/gravitational/teleport/lib/backend"
	"github.com/gravitational/teleport/lib/backend/memory"
	"github.com/gravitational/teleport/lib/cache"
	"github.com/gravitational/teleport/lib/cryptosuites"
	"github.com/gravitational/teleport/lib/defaults"
	"github.com/gravitational/teleport/lib/events"
	"github.com/gravitational/teleport/lib/events/eventstest"
	"github.com/gravitational/teleport/lib/fixtures"
	"github.com/gravitational/teleport/lib/limiter"
	"github.com/gravitational/teleport/lib/service/servicecfg"
	"github.com/gravitational/teleport/lib/services"
	"github.com/gravitational/teleport/lib/services/local"
	"github.com/gravitational/teleport/lib/tlsca"
	"github.com/gravitational/teleport/lib/utils"
)

// AuthServerConfig is auth server test config
type AuthServerConfig struct {
	// ClusterName is cluster name
	ClusterName string
	// ClusterID is the cluster ID; optional - sets to random UUID string if not present
	ClusterID string
	// Dir is directory for local backend
	Dir string
	// AcceptedUsage is an optional list of restricted
	// server usage
	AcceptedUsage []string
	// CipherSuites is the list of ciphers that the server supports.
	CipherSuites []uint16
	// Clock is used to control time in tests.
	Clock clockwork.Clock
	// ClusterNetworkingConfig allows a test to change the default
	// networking configuration.
	ClusterNetworkingConfig types.ClusterNetworkingConfig
	// Streamer allows a test to set its own session recording streamer.
	Streamer events.Streamer
	// AuditLog allows a test to configure its own audit log.
	AuditLog events.AuditLogSessionStreamer
	// TraceClient allows a test to configure the trace client
	TraceClient otlptrace.Client
	// AuthPreferenceSpec is custom initial AuthPreference spec for the test.
	AuthPreferenceSpec *types.AuthPreferenceSpecV2
	// CacheEnabled enables the primary auth server cache.
	CacheEnabled bool
	// RunWhileLockedRetryInterval is the interval to retry the run while locked
	// operation.
	RunWhileLockedRetryInterval time.Duration
	// FIPS means the cluster should run in FIPS mode.
	FIPS bool
	// KeystoreConfig is configuration for the CA keystore.
	KeystoreConfig servicecfg.KeystoreConfig
}

// CheckAndSetDefaults checks and sets defaults
func (cfg *AuthServerConfig) CheckAndSetDefaults() error {
	if cfg.ClusterName == "" {
		cfg.ClusterName = "localhost"
	}
	if cfg.Dir == "" {
		return trace.BadParameter("missing parameter Dir")
	}
	if cfg.Clock == nil {
		cfg.Clock = clockwork.NewFakeClock()
	}
	if len(cfg.CipherSuites) == 0 {
		cfg.CipherSuites = utils.DefaultCipherSuites()
	}
	if cfg.AuthPreferenceSpec == nil {
		cfg.AuthPreferenceSpec = &types.AuthPreferenceSpecV2{
			Type:         constants.Local,
			SecondFactor: constants.SecondFactorOff,
		}
	}
	return nil
}

// Server defines the set of server components for a test
type Server struct {
	TLS        *TLSServer
	AuthServer *AuthServer
}

// ServerConfig defines the configuration for all server components
type ServerConfig struct {
	// Auth specifies the auth server configuration
	Auth AuthServerConfig
	// TLS optionally specifies the configuration for the TLS server.
	// If unspecified, will be generated automatically
	TLS *TLSServerConfig
}

// NewTestServer creates a new test server configuration
func NewTestServer(cfg ServerConfig) (*Server, error) {
	authServer, err := NewAuthServer(cfg.Auth)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	// Set the (test) auth server in cfg.TLS and set any defaults that
	// are not set.
	tlsCfg := cfg.TLS
	if tlsCfg == nil {
		tlsCfg = &TLSServerConfig{}
	}
	if tlsCfg.APIConfig == nil {
		tlsCfg.APIConfig = &auth.APIConfig{}
	}

	tlsCfg.AuthServer = authServer
	tlsCfg.APIConfig.AuthServer = authServer.AuthServer

	if tlsCfg.APIConfig.Authorizer == nil {
		tlsCfg.APIConfig.Authorizer = authServer.Authorizer
	}
	if tlsCfg.APIConfig.AuditLog == nil {
		tlsCfg.APIConfig.AuditLog = authServer.AuditLog
	}
	if tlsCfg.APIConfig.Emitter == nil {
		tlsCfg.APIConfig.Emitter = authServer.AuthServer
	}
	if tlsCfg.AcceptedUsage == nil {
		tlsCfg.AcceptedUsage = authServer.AcceptedUsage
	}

	tlsServer, err := NewTestTLSServer(*tlsCfg)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &Server{
		AuthServer: authServer,
		TLS:        tlsServer,
	}, nil
}

// Auth returns the underlying auth server instance
func (a *Server) Auth() *auth.Server {
	return a.AuthServer.AuthServer
}

func (a *Server) NewClient(identity TestIdentity) (*authclient.Client, error) {
	return a.TLS.NewClient(identity)
}

func (a *Server) ClusterName() string {
	return a.TLS.ClusterName()
}

// Shutdown stops this server instance gracefully
func (a *Server) Shutdown(ctx context.Context) error {
	return trace.NewAggregate(
		a.TLS.Shutdown(ctx),
		a.AuthServer.Close(),
	)
}

// WithClock is a functional server option that sets the server's clock
func WithClock(clock clockwork.Clock) auth.ServerOption {
	return func(s *auth.Server) error {
		s.SetClock(clock)
		return nil
	}
}

// WithBcryptCost is a functional server option that sets the server's bcrypt cost.
func WithBcryptCost(cost int) auth.ServerOption {
	return func(s *auth.Server) error {
		s.SetBcryptCost(cost)
		return nil
	}
}

// AuthServer is an auth server using local filesystem backend
// and test certificate authority key generation that speeds up
// keygen by using the same private key
type AuthServer struct {
	// AuthServer config is configuration used for auth server setup
	AuthServerConfig
	// AuthServer is an auth server
	AuthServer *auth.Server
	// AuditLog is an event audit log
	AuditLog events.AuditLogSessionStreamer
	// Backend is a backend for auth server
	Backend backend.Backend
	// Authorizer is an authorizer used in tests
	Authorizer authz.Authorizer
	// LockWatcher is a lock watcher used in tests.
	LockWatcher *services.LockWatcher
}

// NewAuthServer returns new instances of Auth server
func NewAuthServer(cfg AuthServerConfig) (*AuthServer, error) {
	ctx := context.Background()

	if err := cfg.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}
	srv := &AuthServer{
		AuthServerConfig: cfg,
	}
	b, err := memory.New(memory.Config{
		Context:   ctx,
		Clock:     cfg.Clock,
		EventsOff: false,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// Wrap backend in sanitizer like in production.
	srv.Backend = backend.NewSanitizer(b)

	if cfg.AuditLog != nil {
		srv.AuditLog = cfg.AuditLog
	} else {
		localLog, err := events.NewAuditLog(events.AuditLogConfig{
			DataDir:       cfg.Dir,
			ServerID:      cfg.ClusterName,
			Clock:         cfg.Clock,
			UploadHandler: eventstest.NewMemoryUploader(),
		})
		if err != nil {
			return nil, trace.Wrap(err)
		}
		srv.AuditLog = localLog
	}

	access := local.NewAccessService(srv.Backend)
	identity, err := local.NewTestIdentityService(srv.Backend)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	emitter, err := events.NewCheckingEmitter(events.CheckingEmitterConfig{
		Inner: srv.AuditLog,
		Clock: cfg.Clock,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	accessLists, err := local.NewAccessListService(srv.Backend, cfg.Clock, local.WithRunWhileLockedRetryInterval(cfg.RunWhileLockedRetryInterval))
	if err != nil {
		return nil, trace.Wrap(err)
	}

	clusterName, err := services.NewClusterNameWithRandomID(types.ClusterNameSpecV2{
		ClusterName: cfg.ClusterName,
		ClusterID:   cfg.ClusterID,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	srv.AuthServer, err = auth.NewServer(&auth.InitConfig{
		DataDir:                cfg.Dir,
		Backend:                srv.Backend,
		VersionStorage:         NewFakeTeleportVersion(),
		Authority:              authority.NewWithClock(cfg.Clock),
		Access:                 access,
		Identity:               identity,
		AuditLog:               srv.AuditLog,
		Streamer:               cfg.Streamer,
		SkipPeriodicOperations: true,
		Emitter:                emitter,
		TraceClient:            cfg.TraceClient,
		Clock:                  cfg.Clock,
		ClusterName:            clusterName,
		HostUUID:               uuid.New().String(),
		AccessLists:            accessLists,
		FIPS:                   cfg.FIPS,
		KeyStoreConfig:         cfg.KeystoreConfig,
	},
		WithClock(cfg.Clock),
		// Reduce auth.Server bcrypt costs when testing.
		WithBcryptCost(bcrypt.MinCost),
	)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if cfg.CacheEnabled {
		if err := InitAuthCache(AuthCacheParams{
			AuthServer: srv.AuthServer,
			Unstarted:  true,
		}); err != nil {
			return nil, trace.Wrap(err)
		}
	}

	err = srv.AuthServer.SetClusterAuditConfig(ctx, types.DefaultClusterAuditConfig())
	if err != nil {
		return nil, trace.Wrap(err)
	}

	clusterNetworkingCfg := cfg.ClusterNetworkingConfig
	if clusterNetworkingCfg == nil {
		clusterNetworkingCfg = types.DefaultClusterNetworkingConfig()
	}

	_, err = srv.AuthServer.UpsertClusterNetworkingConfig(ctx, clusterNetworkingCfg)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	_, err = srv.AuthServer.UpsertSessionRecordingConfig(ctx, types.DefaultSessionRecordingConfig())
	if err != nil {
		return nil, trace.Wrap(err)
	}

	err = srv.AuthServer.SetClusterName(clusterName)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	authPreference, err := types.NewAuthPreferenceFromConfigFile(*cfg.AuthPreferenceSpec)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	if authPreference.GetSignatureAlgorithmSuite() == types.SignatureAlgorithmSuite_SIGNATURE_ALGORITHM_SUITE_UNSPECIFIED {
		authPreference.SetSignatureAlgorithmSuite(types.SignatureAlgorithmSuite_SIGNATURE_ALGORITHM_SUITE_BALANCED_V1)
	}
	_, err = srv.AuthServer.UpsertAuthPreference(ctx, authPreference)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	token, err := types.NewProvisionTokenFromSpec("static-token", time.Unix(0, 0).UTC(), types.ProvisionTokenSpecV2{
		Roles: types.SystemRoles{types.RoleNode},
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	// set static tokens
	staticTokens, err := types.NewStaticTokens(types.StaticTokensSpecV2{
		StaticTokens: []types.ProvisionTokenV1{
			*token.V1(),
		},
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	err = srv.AuthServer.SetStaticTokens(staticTokens)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// create the default role
	if _, err := srv.AuthServer.UpsertRole(ctx, services.NewImplicitRole()); err != nil {
		return nil, trace.Wrap(err)
	}

	// Setup certificate and signing authorities.
	for _, caType := range types.CertAuthTypes {
		if err = srv.AuthServer.UpsertCertAuthority(ctx, NewTestCAWithConfig(TestCAConfig{
			Type:        caType,
			ClusterName: srv.ClusterName,
			Clock:       cfg.Clock,
		})); err != nil {
			return nil, trace.Wrap(err)
		}
	}

	srv.LockWatcher, err = services.NewLockWatcher(ctx, services.LockWatcherConfig{
		ResourceWatcherConfig: services.ResourceWatcherConfig{
			Component:      teleport.ComponentAuth,
			Client:         srv.AuthServer,
			Clock:          cfg.Clock,
			MaxRetryPeriod: defaults.HighResPollingPeriod,
		},
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	srv.AuthServer.SetLockWatcher(srv.LockWatcher)

	unifiedResourcesCache, err := services.NewUnifiedResourceCache(srv.AuthServer.CloseContext(), services.UnifiedResourceCacheConfig{
		ResourceWatcherConfig: services.ResourceWatcherConfig{
			QueueSize:    defaults.UnifiedResourcesQueueSize,
			Component:    teleport.ComponentUnifiedResource,
			Client:       srv.AuthServer,
			MaxStaleness: time.Minute,
		},
		ResourceGetter: srv.AuthServer,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	srv.AuthServer.SetUnifiedResourcesCache(unifiedResourcesCache)

	accessRequestCache, err := services.NewAccessRequestCache(services.AccessRequestCacheConfig{
		Events: srv.AuthServer.Services,
		Getter: srv.AuthServer.Services,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	srv.AuthServer.SetAccessRequestCache(accessRequestCache)

	headlessAuthenticationWatcher, err := local.NewHeadlessAuthenticationWatcher(srv.AuthServer.CloseContext(), local.HeadlessAuthenticationWatcherConfig{
		Backend: b,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	if err := headlessAuthenticationWatcher.WaitInit(ctx); err != nil {
		return nil, trace.Wrap(err)
	}
	srv.AuthServer.SetHeadlessAuthenticationWatcher(headlessAuthenticationWatcher)

	srv.Authorizer, err = authz.NewAuthorizer(authz.AuthorizerOpts{
		ClusterName:         srv.ClusterName,
		AccessPoint:         srv.AuthServer,
		ReadOnlyAccessPoint: srv.AuthServer.ReadOnlyCache,
		LockWatcher:         srv.LockWatcher,
		// AuthServer does explicit device authorization checks.
		DeviceAuthorization: authz.DeviceAuthorizationOpts{
			DisableGlobalMode: true,
			DisableRoleMode:   true,
		},
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	userNotificationCache, err := services.NewUserNotificationCache(services.NotificationCacheConfig{
		Events: srv.AuthServer.Services,
		Getter: srv.AuthServer.Cache,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	srv.AuthServer.SetUserNotificationCache(userNotificationCache)

	globalNotificationCache, err := services.NewGlobalNotificationCache(services.NotificationCacheConfig{
		Events: srv.AuthServer.Services,
		Getter: srv.AuthServer.Cache,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	srv.AuthServer.SetGlobalNotificationCache(globalNotificationCache)

	// Auth initialization is done (including creation/updating of all singleton
	// configuration resources) so now we can start the cache.
	if c, ok := srv.AuthServer.Cache.(*cache.Cache); ok {
		if err := c.Start(); err != nil {
			return nil, trace.NewAggregate(err, c.Close())
		}
	}
	return srv, nil
}

type AuthCacheParams struct {
	AuthServer *auth.Server
	Unstarted  bool
}

func InitAuthCache(p AuthCacheParams) error {
	c, err := accesspoint.NewCache(accesspoint.Config{
		Context:      p.AuthServer.CloseContext(),
		Setup:        cache.ForAuth,
		CacheName:    []string{teleport.ComponentAuth},
		EventsSystem: true,
		Unstarted:    p.Unstarted,

		Access:                  p.AuthServer.Services.Access,
		AccessLists:             p.AuthServer.Services.AccessLists,
		AccessMonitoringRules:   p.AuthServer.Services.AccessMonitoringRules,
		AppSession:              p.AuthServer.Services.Identity,
		Applications:            p.AuthServer.Services.Applications,
		ClusterConfig:           p.AuthServer.Services.ClusterConfigurationInternal,
		CrownJewels:             p.AuthServer.Services.CrownJewels,
		DatabaseObjects:         p.AuthServer.Services.DatabaseObjects,
		DatabaseServices:        p.AuthServer.Services.DatabaseServices,
		Databases:               p.AuthServer.Services.Databases,
		DiscoveryConfigs:        p.AuthServer.Services.DiscoveryConfigs,
		DynamicAccess:           p.AuthServer.Services.DynamicAccessExt,
		Events:                  p.AuthServer.Services.Events,
		Integrations:            p.AuthServer.Services.Integrations,
		KubeWaitingContainers:   p.AuthServer.Services.KubeWaitingContainer,
		Kubernetes:              p.AuthServer.Services.Kubernetes,
		Notifications:           p.AuthServer.Services.Notifications,
		Okta:                    p.AuthServer.Services.Okta,
		Presence:                p.AuthServer.Services.PresenceInternal,
		Provisioner:             p.AuthServer.Services.Provisioner,
		Restrictions:            p.AuthServer.Services.Restrictions,
		SAMLIdPServiceProviders: p.AuthServer.Services.SAMLIdPServiceProviders,
		SecReports:              p.AuthServer.Services.SecReports,
		SnowflakeSession:        p.AuthServer.Services.Identity,
		SPIFFEFederations:       p.AuthServer.Services.SPIFFEFederations,
		StaticHostUsers:         p.AuthServer.Services.StaticHostUser,
		Trust:                   p.AuthServer.Services.TrustInternal,
		UserGroups:              p.AuthServer.Services.UserGroups,
		UserTasks:               p.AuthServer.Services.UserTasks,
		UserLoginStates:         p.AuthServer.Services.UserLoginStates,
		Users:                   p.AuthServer.Services.Identity,
		WebSession:              p.AuthServer.Services.Identity.WebSessions(),
		WebToken:                p.AuthServer.Services.WebTokens(),
		WorkloadIdentity:        p.AuthServer.Services.WorkloadIdentities,
		DynamicWindowsDesktops:  p.AuthServer.Services.DynamicWindowsDesktops,
		WindowsDesktops:         p.AuthServer.Services.WindowsDesktops,
		AutoUpdateService:       p.AuthServer.Services.AutoUpdateService,
		ProvisioningStates:      p.AuthServer.Services.ProvisioningStates,
		IdentityCenter:          p.AuthServer.Services.IdentityCenter,
		PluginStaticCredentials: p.AuthServer.Services.PluginStaticCredentials,
		GitServers:              p.AuthServer.Services.GitServers,
		HealthCheckConfig:       p.AuthServer.Services.HealthCheckConfig,
		BotInstance:             p.AuthServer.Services.BotInstance,
		RecordingEncryption:     p.AuthServer.Services.RecordingEncryptionManager,
		Plugin:                  p.AuthServer.Services.Plugins,
	})
	if err != nil {
		return trace.Wrap(err)
	}
	p.AuthServer.Cache = c
	return nil
}

func (a *AuthServer) Close() error {
	defer a.LockWatcher.Close()

	return trace.NewAggregate(
		a.AuthServer.Close(),
		a.Backend.Close(),
		a.AuditLog.Close(),
	)
}

// GenerateUserCert takes the public key in the OpenSSH `authorized_keys`
// plain text format, signs it using User Certificate Authority signing key and returns the
// resulting certificate.
func (a *AuthServer) GenerateUserCert(key []byte, username string, ttl time.Duration, compatibility string) ([]byte, error) {
	sshCert, _, err := a.AuthServer.GenerateUserTestCerts(auth.GenerateUserTestCertsRequest{
		SSHPubKey:     key,
		Username:      username,
		TTL:           ttl,
		Compatibility: compatibility,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return sshCert, nil
}

// PrivateKeyToPublicKeyTLS gets the TLS public key from a raw private key.
func PrivateKeyToPublicKeyTLS(privateKey []byte) (tlsPublicKey []byte, err error) {
	sshPrivate, err := ssh.ParseRawPrivateKey(privateKey)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	tlsPublicKey, err = tlsca.MarshalPublicKeyFromPrivateKeyPEM(sshPrivate)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return tlsPublicKey, nil
}

// generateCertificate generates certificate for identity,
// returns private public key pair
func generateCertificate(authServer *auth.Server, identity TestIdentity) ([]byte, []byte, error) {
	ctx := context.TODO()

	key, err := cryptosuites.GenerateKeyWithAlgorithm(cryptosuites.ECDSAP256)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	privateKeyPEM, err := keys.MarshalPrivateKey(key)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	tlsPublicKeyPEM, err := keys.MarshalPublicKey(key.Public())
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	sshPublicKey, err := ssh.NewPublicKey(key.Public())
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}
	sshPublicKeyPEM := ssh.MarshalAuthorizedKey(sshPublicKey)

	switch id := identity.I.(type) {
	case authz.LocalUser:
		if identity.TTL == 0 {
			identity.TTL = time.Hour
		}

		_, tlsCert, err := authServer.GenerateUserTestCerts(auth.GenerateUserTestCertsRequest{
			SSHPubKey:        sshPublicKeyPEM,
			TLSPubKey:        tlsPublicKeyPEM,
			Username:         id.Username,
			TTL:              identity.TTL,
			RouteToCluster:   identity.RouteToCluster,
			PinnedIP:         id.Identity.PinnedIP,
			MFAVerified:      id.Identity.MFAVerified,
			DeviceExtensions: auth.DeviceExtensions(id.Identity.DeviceExtensions),
			Generation:       id.Identity.Generation,
			Renewable:        identity.Renewable,
			Usage:            identity.AcceptedUsage,
		})
		if err != nil {
			return nil, nil, trace.Wrap(err)
		}

		return tlsCert, privateKeyPEM, nil
	case authz.BuiltinRole:
		certs, err := authServer.GenerateHostCerts(ctx,
			&proto.HostCertsRequest{
				HostID:       id.Username,
				NodeName:     id.Username,
				Role:         id.Role,
				PublicTLSKey: tlsPublicKeyPEM,
				PublicSSHKey: sshPublicKeyPEM,
				SystemRoles:  id.AdditionalSystemRoles,
			})
		if err != nil {
			return nil, nil, trace.Wrap(err)
		}
		return certs.TLS, privateKeyPEM, nil
	case authz.RemoteBuiltinRole:
		certs, err := authServer.GenerateHostCerts(ctx,
			&proto.HostCertsRequest{
				HostID:       id.Username,
				NodeName:     id.Username,
				Role:         id.Role,
				PublicTLSKey: tlsPublicKeyPEM,
				PublicSSHKey: sshPublicKeyPEM,
			})
		if err != nil {
			return nil, nil, trace.Wrap(err)
		}
		return certs.TLS, privateKeyPEM, nil
	default:
		return nil, nil, trace.BadParameter("identity of unknown type %T is unsupported", identity)
	}
}

// NewCertificate returns new TLS credentials generated by test auth server
func (a *AuthServer) NewCertificate(identity TestIdentity) (*tls.Certificate, error) {
	cert, key, err := generateCertificate(a.AuthServer, identity)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &tlsCert, nil
}

// Clock returns clock used by auth server
func (a *AuthServer) Clock() clockwork.Clock {
	return a.AuthServer.GetClock()
}

// Trust adds other server host certificate authority as trusted
func (a *AuthServer) Trust(ctx context.Context, remote *AuthServer, roleMap types.RoleMap) error {
	remoteCA, err := remote.AuthServer.GetCertAuthority(ctx, types.CertAuthID{
		Type:       types.HostCA,
		DomainName: remote.ClusterName,
	}, false)
	if err != nil {
		return trace.Wrap(err)
	}
	err = a.AuthServer.UpsertCertAuthority(ctx, remoteCA)
	if err != nil {
		return trace.Wrap(err)
	}
	remoteCA, err = remote.AuthServer.GetCertAuthority(ctx, types.CertAuthID{
		Type:       types.DatabaseCA,
		DomainName: remote.ClusterName,
	}, false)
	if err != nil {
		return trace.Wrap(err)
	}
	err = a.AuthServer.UpsertCertAuthority(ctx, remoteCA)
	if err != nil {
		return trace.Wrap(err)
	}
	remoteCA, err = remote.AuthServer.GetCertAuthority(ctx, types.CertAuthID{
		Type:       types.DatabaseClientCA,
		DomainName: remote.ClusterName,
	}, false)
	if err != nil {
		return trace.Wrap(err)
	}
	err = a.AuthServer.UpsertCertAuthority(ctx, remoteCA)
	if err != nil {
		return trace.Wrap(err)
	}
	remoteCA, err = remote.AuthServer.GetCertAuthority(ctx, types.CertAuthID{
		Type:       types.OpenSSHCA,
		DomainName: remote.ClusterName,
	}, false)
	if err != nil {
		return trace.Wrap(err)
	}
	err = a.AuthServer.UpsertCertAuthority(ctx, remoteCA)
	if err != nil {
		return trace.Wrap(err)
	}
	remoteCA, err = remote.AuthServer.GetCertAuthority(ctx, types.CertAuthID{
		Type:       types.UserCA,
		DomainName: remote.ClusterName,
	}, false)
	if err != nil {
		return trace.Wrap(err)
	}
	remoteCA.SetRoleMap(roleMap)
	err = a.AuthServer.UpsertCertAuthority(ctx, remoteCA)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// NewTestTLSServer returns new test TLS server
func (a *AuthServer) NewTestTLSServer(opts ...TestTLSServerOption) (*TLSServer, error) {
	apiConfig := &auth.APIConfig{
		AuthServer: a.AuthServer,
		Authorizer: a.Authorizer,
		AuditLog:   a.AuditLog,
		Emitter:    a.AuthServer,
	}
	cfg := TLSServerConfig{
		APIConfig:     apiConfig,
		AuthServer:    a,
		AcceptedUsage: a.AcceptedUsage,
	}
	for _, o := range opts {
		o(&cfg)
	}
	srv, err := NewTestTLSServer(cfg)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return srv, nil
}

// TestTLSServerOption is a functional option passed to NewTestTLSServer
type TestTLSServerOption func(*TLSServerConfig)

// WithLimiterConfig sets connection and request limiter configuration.
func WithLimiterConfig(config *limiter.Config) TestTLSServerOption {
	return func(cfg *TLSServerConfig) {
		cfg.Limiter = config
	}
}

// WithAccessGraphConfig sets the access graph configuration.
func WithAccessGraphConfig(config auth.AccessGraphConfig) TestTLSServerOption {
	return func(cfg *TLSServerConfig) {
		cfg.APIConfig.AccessGraph = config
	}
}

// NewRemoteClient creates new client to the remote server using identity
// generated for this certificate authority
func (a *AuthServer) NewRemoteClient(identity TestIdentity, addr net.Addr, pool *x509.CertPool) (*authclient.Client, error) {
	tlsConfig := utils.TLSConfig(a.CipherSuites)
	cert, err := a.NewCertificate(identity)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	tlsConfig.Certificates = []tls.Certificate{*cert}
	tlsConfig.RootCAs = pool
	tlsConfig.ServerName = apiutils.EncodeClusterName(a.ClusterName)
	tlsConfig.Time = a.AuthServer.GetClock().Now

	return authclient.NewClient(client.Config{
		Addrs: []string{addr.String()},
		Credentials: []client.Credentials{
			client.LoadTLS(tlsConfig),
		},
		CircuitBreakerConfig: breaker.NoopBreakerConfig(),
	})
}

// TLSServerConfig is a configuration for test TLS server
type TLSServerConfig struct {
	// APIConfig is a configuration of API server
	APIConfig *auth.APIConfig
	// AuthServer is a test auth server used to serve requests
	AuthServer *AuthServer
	// Limiter is a connection and request limiter
	Limiter *limiter.Config
	// Listener is a listener to serve requests on
	Listener net.Listener
	// AcceptedUsage is a list of accepted usage restrictions
	AcceptedUsage []string
}

// Auth returns auth server used by this TLS server
func (t *TLSServer) Auth() *auth.Server {
	return t.AuthServer.AuthServer
}

// TLSServer is a test TLS server
type TLSServer struct {
	// TLSServerConfig is a configuration for TLS server
	TLSServerConfig
	// Identity is a generated TLS/SSH identity used to answer in TLS
	Identity *state.Identity
	// TLSServer is a configured TLS server
	TLSServer *auth.TLSServer
}

// ClusterName returns name of test TLS server cluster
func (t *TLSServer) ClusterName() string {
	return t.AuthServer.ClusterName
}

// Clock returns clock used by auth server
func (t *TLSServer) Clock() clockwork.Clock {
	return t.AuthServer.Clock()
}

// CheckAndSetDefaults checks and sets limiter defaults
func (cfg *TLSServerConfig) CheckAndSetDefaults() error {
	if cfg.APIConfig == nil {
		return trace.BadParameter("missing parameter APIConfig")
	}
	if cfg.AuthServer == nil {
		return trace.BadParameter("missing parameter AuthServer")
	}
	// use very permissive limiter configuration by default
	if cfg.Limiter == nil {
		cfg.Limiter = &limiter.Config{
			MaxConnections: 1000,
		}
	}
	return nil
}

// NewTestTLSServer returns new test TLS server that is started and is listening
// on 127.0.0.1 loopback on any available port
func NewTestTLSServer(cfg TLSServerConfig) (*TLSServer, error) {
	err := cfg.CheckAndSetDefaults()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	srv := &TLSServer{
		TLSServerConfig: cfg,
	}
	srv.Identity, err = NewServerIdentity(srv.AuthServer.AuthServer, "test-tls-server", types.RoleAuth)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// Register TLS endpoint of the auth service
	tlsConfig, err := srv.Identity.TLSConfig(srv.AuthServer.CipherSuites)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	tlsConfig.Time = cfg.AuthServer.Clock().Now
	tlsCert := tlsConfig.Certificates[0]

	srv.Listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, trace.Wrap(err)
	}

	srv.TLSServer, err = auth.NewTLSServer(context.Background(), auth.TLSServerConfig{
		Listener:             srv.Listener,
		AccessPoint:          srv.AuthServer.AuthServer.Cache,
		TLS:                  tlsConfig,
		GetClientCertificate: func() (*tls.Certificate, error) { return &tlsCert, nil },
		APIConfig:            *srv.APIConfig,
		LimiterConfig:        *srv.Limiter,
		AcceptedUsage:        cfg.AcceptedUsage,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	if err := srv.Start(); err != nil {
		return nil, trace.Wrap(err)
	}
	return srv, nil
}

// TestIdentity is test identity spec used to generate identities in tests
type TestIdentity struct {
	I              authz.IdentityGetter
	TTL            time.Duration
	AcceptedUsage  []string
	RouteToCluster string
	Renewable      bool
	Generation     uint64
}

// TestUser returns TestIdentity for local user
func TestUser(username string) TestIdentity {
	return TestIdentity{
		I: authz.LocalUser{
			Username: username,
			Identity: tlsca.Identity{Username: username},
		},
	}
}

// TestUserWithDeviceExtensions returns a TestIdentity for a local user,
// including the supplied device extensions in the tlsca.Identity.
func TestUserWithDeviceExtensions(username string, exts tlsca.DeviceExtensions) TestIdentity {
	return TestIdentity{
		I: authz.LocalUser{
			Username: username,
			Identity: tlsca.Identity{
				Username:         username,
				DeviceExtensions: exts,
			},
		},
	}
}

// TestRenewableUser returns a TestIdentity for a local user
// with renewable credentials.
func TestRenewableUser(username string, generation uint64) TestIdentity {
	return TestIdentity{
		I: authz.LocalUser{
			Username: username,
			Identity: tlsca.Identity{
				Username: username,
			},
		},
		Renewable:  true,
		Generation: generation,
	}
}

// TestNop returns "Nop" - unauthenticated identity
func TestNop() TestIdentity {
	return TestIdentity{
		I: nil,
	}
}

// TestAdmin returns TestIdentity for admin user
func TestAdmin() TestIdentity {
	return TestBuiltin(types.RoleAdmin)
}

// TestBuiltin returns TestIdentity for builtin user
func TestBuiltin(role types.SystemRole) TestIdentity {
	return TestIdentity{
		I: authz.BuiltinRole{
			Role:     role,
			Username: string(role),
		},
	}
}

// TestServerID returns a TestIdentity for a node with the passed in serverID.
func TestServerID(role types.SystemRole, serverID string) TestIdentity {
	return TestIdentity{
		I: authz.BuiltinRole{
			Role:                  types.RoleInstance,
			Username:              serverID,
			AdditionalSystemRoles: types.SystemRoles{role},
			Identity: tlsca.Identity{
				Username: serverID,
			},
		},
	}
}

// TestRemoteBuiltin returns TestIdentity for a remote builtin role.
func TestRemoteBuiltin(role types.SystemRole, remoteCluster string) TestIdentity {
	return TestIdentity{
		I: authz.RemoteBuiltinRole{
			Role:        role,
			Username:    string(role),
			ClusterName: remoteCluster,
		},
	}
}

func (i TestIdentity) GetUsername() string {
	return i.I.GetIdentity().Username
}

// NewClientFromWebSession returns new authenticated client from web session
func (t *TLSServer) NewClientFromWebSession(sess types.WebSession) (*authclient.Client, error) {
	tlsConfig, err := t.Identity.TLSConfig(t.AuthServer.CipherSuites)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	tlsCert, err := tls.X509KeyPair(sess.GetTLSCert(), sess.GetTLSPriv())
	if err != nil {
		return nil, trace.Wrap(err, "failed to parse TLS cert and key")
	}
	tlsConfig.Certificates = []tls.Certificate{tlsCert}
	tlsConfig.Time = t.AuthServer.AuthServer.GetClock().Now

	return authclient.NewClient(client.Config{
		Addrs: []string{t.Addr().String()},
		Credentials: []client.Credentials{
			client.LoadTLS(tlsConfig),
		},
		CircuitBreakerConfig: breaker.NoopBreakerConfig(),
	})
}

// CertPool returns cert pool that auth server represents
func (t *TLSServer) CertPool() (*x509.CertPool, error) {
	tlsConfig, err := t.Identity.TLSConfig(t.AuthServer.CipherSuites)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return tlsConfig.RootCAs, nil
}

// ClientTLSConfig returns client TLS config based on the identity
func (t *TLSServer) ClientTLSConfig(identity TestIdentity) (*tls.Config, error) {
	tlsConfig, err := t.Identity.TLSConfig(t.AuthServer.CipherSuites)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	if identity.I != nil {
		cert, err := t.AuthServer.NewCertificate(identity)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		tlsConfig.Certificates = []tls.Certificate{*cert}
	} else {
		// this client is not authenticated, which means that auth
		// server should apply Nop builtin role
		tlsConfig.Certificates = nil
	}
	tlsConfig.Time = t.AuthServer.AuthServer.GetClock().Now
	return tlsConfig, nil
}

// CloneClient uses the same credentials as the passed client
// but forces the client to be recreated
func (t *TLSServer) CloneClient(tt *testing.T, clt *authclient.Client) *authclient.Client {
	tt.Helper()

	tlsConfig := clt.Config()
	// When cloning a client, we want to make sure that we don't reuse
	// the same session ticket cache. The session ticket cache should not be
	// shared between all clients that use the same TLS config.
	// Reusing the cache will skip the TLS handshake and may introduce a weird
	// behavior in tests.
	if tlsConfig.ClientSessionCache != nil {
		tlsConfig = tlsConfig.Clone()
		tlsConfig.ClientSessionCache = tls.NewLRUClientSessionCache(utils.DefaultLRUCapacity)
	}

	newClient, err := authclient.NewClient(client.Config{
		Addrs: []string{t.Addr().String()},
		Credentials: []client.Credentials{
			client.LoadTLS(tlsConfig),
		},
		CircuitBreakerConfig: breaker.NoopBreakerConfig(),
	})
	if err != nil {
		tt.Fatalf("error creating auth client: %v", err.Error())
	}

	tt.Cleanup(func() { _ = newClient.Close() })
	return newClient
}

// NewClientWithCert creates a new client using given cert and private key
func (t *TLSServer) NewClientWithCert(clientCert tls.Certificate) (*authclient.Client, error) {
	tlsConfig, err := t.Identity.TLSConfig(t.AuthServer.CipherSuites)
	if err != nil {
		return nil, err
	}
	tlsConfig.Time = t.AuthServer.AuthServer.GetClock().Now
	tlsConfig.Certificates = []tls.Certificate{clientCert}
	newClient, err := authclient.NewClient(client.Config{
		Addrs: []string{t.Addr().String()},
		Credentials: []client.Credentials{
			client.LoadTLS(tlsConfig),
		},
		CircuitBreakerConfig: breaker.NoopBreakerConfig(),
	})
	if err != nil {
		return nil, err
	}
	return newClient, nil
}

// NewClient returns new client to test server authenticated with identity
func (t *TLSServer) NewClient(identity TestIdentity) (*authclient.Client, error) {
	tlsConfig, err := t.ClientTLSConfig(identity)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	newClient, err := authclient.NewClient(client.Config{
		DialInBackground: true,
		Addrs:            []string{t.Addr().String()},
		Credentials: []client.Credentials{
			client.LoadTLS(tlsConfig),
		},
		CircuitBreakerConfig: breaker.NoopBreakerConfig(),
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return newClient, nil
}

// Addr returns address of TLS server
func (t *TLSServer) Addr() net.Addr {
	return t.Listener.Addr()
}

// Start starts TLS server on loopback address on the first listening socket
func (t *TLSServer) Start() error {
	go t.TLSServer.Serve()
	return nil
}

// Close closes the listener and HTTP server
func (t *TLSServer) Close() error {
	var errs []error
	if err := t.Stop(); err != nil {
		errs = append(errs, err)
	}

	if t.AuthServer.Backend != nil {
		if err := t.AuthServer.Backend.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return trace.NewAggregate(errs...)
}

// Shutdown closes the listener and HTTP server gracefully
func (t *TLSServer) Shutdown(ctx context.Context) error {
	var errs []error
	if err := t.TLSServer.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}

	if t.Listener != nil {
		if err := t.Listener.Close(); err != nil && !utils.IsUseOfClosedNetworkError(err) {
			errs = append(errs, err)
		}

	}
	if t.AuthServer.Backend != nil {
		if err := t.AuthServer.Backend.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return trace.NewAggregate(errs...)
}

// Stop stops listening server, but does not close the auth backend
func (t *TLSServer) Stop() error {
	var errs []error
	if err := t.TLSServer.Close(); err != nil {
		errs = append(errs, err)
	}
	if t.Listener != nil {
		if err := t.Listener.Close(); err != nil && !utils.IsUseOfClosedNetworkError(err) {
			errs = append(errs, err)
		}
	}

	return trace.NewAggregate(errs...)
}

// FakeTeleportVersion fake version storage implementation always return current version.
type FakeTeleportVersion struct{}

// NewFakeTeleportVersion creates fake version storage.
func NewFakeTeleportVersion() *FakeTeleportVersion {
	return &FakeTeleportVersion{}
}

// GetTeleportVersion returns current Teleport version.
func (s FakeTeleportVersion) GetTeleportVersion(_ context.Context) (semver.Version, error) {
	return *teleport.SemVer(), nil
}

// WriteTeleportVersion stub function for writing.
func (s FakeTeleportVersion) WriteTeleportVersion(_ context.Context, _ semver.Version) error {
	return nil
}

// DeleteTeleportVersion error stub function for deleting.
func (s FakeTeleportVersion) DeleteTeleportVersion(_ context.Context) error {
	return nil
}

// NewServerIdentity generates new server identity, used in tests
func NewServerIdentity(clt *auth.Server, hostID string, role types.SystemRole) (*state.Identity, error) {
	key, err := cryptosuites.GenerateKeyWithAlgorithm(cryptosuites.ECDSAP256)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	privateKeyPEM, err := keys.MarshalPrivateKey(key)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	sshPubKey, err := ssh.NewPublicKey(key.Public())
	if err != nil {
		return nil, trace.Wrap(err)
	}

	tlsPubKey, err := keys.MarshalPublicKey(key.Public())
	if err != nil {
		return nil, trace.Wrap(err)
	}

	certs, err := clt.GenerateHostCerts(context.Background(),
		&proto.HostCertsRequest{
			HostID:       hostID,
			NodeName:     hostID,
			Role:         role,
			PublicSSHKey: ssh.MarshalAuthorizedKey(sshPubKey),
			PublicTLSKey: tlsPubKey,
		})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return state.ReadIdentityFromKeyPair(privateKeyPEM, certs)
}

// UserClient allows overwriting a user.
type UserClient interface {
	UpsertUser(context.Context, types.User) (types.User, error)
}

// RoleClient allows overwriting a role.
type RoleClient interface {
	UpsertRole(context.Context, types.Role) (types.Role, error)
}

// UserRoleClient allows overwriting uesrs and roles.
type UserRoleClient interface {
	UserClient
	RoleClient
}

// CreateRole creates a role without assigning any users. Used in tests.
func CreateRole(ctx context.Context, clt RoleClient, name string, spec types.RoleSpecV6) (types.Role, error) {
	role, err := types.NewRole(name, spec)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	upserted, err := clt.UpsertRole(ctx, role)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return upserted, nil
}

// CreateUserRoleAndRequestable creates two roles for a user, one base role with allowed login
// matching username, and another role with a login matching rolename that can be requested.
func CreateUserRoleAndRequestable(clt UserRoleClient, username string, rolename string) (types.User, error) {
	ctx := context.TODO()
	user, err := types.NewUser(username)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	baseRole := services.RoleForUser(user)
	baseRole.SetAccessRequestConditions(types.Allow, types.AccessRequestConditions{
		Roles: []string{rolename},
	})
	baseRole.SetLogins(types.Allow, nil)
	baseRole, err = clt.UpsertRole(ctx, baseRole)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	user.AddRole(baseRole.GetName())
	_, err = clt.UpsertUser(ctx, user)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	requestableRole := services.RoleForUser(user)
	requestableRole.SetName(rolename)
	requestableRole.SetLogins(types.Allow, []string{rolename})
	_, err = clt.UpsertRole(ctx, requestableRole)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return user, nil
}

// CreateAccessPluginUser creates a user with list/read abilites for access requests, and list/read/update
// abilities for access plugin data.
func CreateAccessPluginUser(ctx context.Context, clt UserRoleClient, username string) (types.User, error) {
	user, err := types.NewUser(username)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	role := services.RoleForUser(user)
	rules := role.GetRules(types.Allow)
	rules = append(rules,
		types.Rule{
			Resources: []string{types.KindAccessRequest},
			Verbs:     []string{types.VerbRead, types.VerbList},
		},
		types.Rule{
			Resources: []string{types.KindAccessPluginData},
			Verbs:     []string{types.VerbRead, types.VerbList, types.VerbUpdate},
		},
	)
	role.SetRules(types.Allow, rules)
	role.SetLogins(types.Allow, nil)
	upsertedRole, err := clt.UpsertRole(ctx, role)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	user.AddRole(upsertedRole.GetName())
	if _, err := clt.UpsertUser(ctx, user); err != nil {
		return nil, trace.Wrap(err)
	}
	return user, nil
}

// CreateUser creates user and role and assigns role to a user, used in tests
func CreateUser(ctx context.Context, clt UserRoleClient, username string, roles ...types.Role) (types.User, error) {
	user, err := types.NewUser(username)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	for _, role := range roles {
		upsertedRole, err := clt.UpsertRole(ctx, role)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		user.AddRole(upsertedRole.GetName())
	}

	created, err := clt.UpsertUser(ctx, user)
	return created, trace.Wrap(err)
}

// createUserAndRoleOptions is a set of options for CreateUserAndRole
type createUserAndRoleOptions struct {
	mutateUser []func(user types.User)
	mutateRole []func(role types.Role)
	version    string
}

// CreateUserAndRoleOption is a functional option for CreateUserAndRole
type CreateUserAndRoleOption func(*createUserAndRoleOptions)

// WithRoleVersion sets the version of the role to be created.
func WithRoleVersion(version string) CreateUserAndRoleOption {
	return func(o *createUserAndRoleOptions) {
		o.version = version
	}
}

// WithUserMutator sets a function that will be called to mutate the user before it is created
func WithUserMutator(mutate ...func(user types.User)) CreateUserAndRoleOption {
	return func(o *createUserAndRoleOptions) {
		o.mutateUser = append(o.mutateUser, mutate...)
	}
}

// WithRoleMutator sets a function that will be called to mutate the role before it is created
func WithRoleMutator(mutate ...func(role types.Role)) CreateUserAndRoleOption {
	return func(o *createUserAndRoleOptions) {
		o.mutateRole = append(o.mutateRole, mutate...)
	}
}

// CreateUserAndRole creates user and role and assigns role to a user, used in tests
// If allowRules is nil, the role has admin privileges.
// If allowRules is not-nil, then the rules associated with the role will be
// replaced with those specified.
func CreateUserAndRole(clt UserRoleClient, username string, allowedLogins []string, allowRules []types.Rule, opts ...CreateUserAndRoleOption) (types.User, types.Role, error) {
	o := createUserAndRoleOptions{
		version: types.DefaultRoleVersion,
	}
	for _, opt := range opts {
		opt(&o)
	}
	ctx := context.TODO()
	user, err := types.NewUser(username)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	role := services.RoleWithVersionForUser(user, o.version)
	role.SetLogins(types.Allow, allowedLogins)
	if allowRules != nil {
		role.SetRules(types.Allow, allowRules)
	}
	for _, mutate := range o.mutateRole {
		mutate(role)
	}
	upsertedRole, err := clt.UpsertRole(ctx, role)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	user.AddRole(upsertedRole.GetName())
	for _, mutate := range o.mutateUser {
		mutate(user)
	}
	created, err := clt.UpsertUser(ctx, user)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}
	return created, role, nil
}

// CreateUserAndRoleWithoutRoles creates user and role, but does not assign user to a role, used in tests
func CreateUserAndRoleWithoutRoles(clt UserRoleClient, username string, allowedLogins []string) (types.User, types.Role, error) {
	ctx := context.TODO()
	user, err := types.NewUser(username)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	role := services.RoleForUser(user)
	set := services.MakeRuleSet(role.GetRules(types.Allow))
	delete(set, types.KindRole)
	role.SetRules(types.Allow, set.Slice())
	role.SetLogins(types.Allow, []string{user.GetName()})
	upsertedRole, err := clt.UpsertRole(ctx, role)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	user.AddRole(upsertedRole.GetName())
	created, err := clt.UpsertUser(ctx, user)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	return created, upsertedRole, nil
}

// Flusher is the set of methods expected by the FlushCache helper.
type Flusher interface {
	// GetRole returns role by name
	GetRole(ctx context.Context, name string) (types.Role, error)
	// CreateRole creates a new role.
	CreateRole(context.Context, types.Role) (types.Role, error)
	// DeleteRole deletes the role by name.
	DeleteRole(ctx context.Context, name string) error
}

// FlushCache is a helper for waiting until preceding changes have propagated to the
// cache during a test. this is useful for writing tests that may want to update backend
// state and then perform some operation that depends on the auth server knowing that state.
// note that this is only intended for use with the memory backend, as this helper relies on the assumption that
// write events for different keys show up in the order in which the writes were performed, which
// is not necessarily true for all backends.
func FlushCache(t *testing.T, clt Flusher) {
	ctx := context.Background()

	// the pattern of writing a resource and then waiting for it to appear
	// works for any resource type (when using memory backend).
	name := strings.ReplaceAll(uuid.NewString(), "-", "")
	defer clt.DeleteRole(ctx, name)

	role, err := types.NewRole(name, types.RoleSpecV6{})
	if err != nil {
		t.Fatalf("Failed to instantiate new role: %v", err)
	}

	role, err = clt.CreateRole(ctx, role)
	if err != nil {
		t.Fatalf("Failed to create new role: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	for {
		r, err := clt.GetRole(ctx, name)
		if err == nil && r.GetRevision() == role.GetRevision() {
			return
		}

		select {
		case <-time.After(200 * time.Millisecond):
		case <-ctx.Done():
			t.Fatal("Time out waiting for role to be replicated")
		}
	}
}

// TestCAConfig defines the configuration for generating
// a test certificate authority
type TestCAConfig struct {
	Type        types.CertAuthType
	PrivateKeys [][]byte
	Clock       clockwork.Clock
	ClusterName string
	// the below string fields default to ClusterName if left empty
	ResourceName        string
	SubjectOrganization string
}

// NewTestCA returns new test authority with a test key as a public and
// signing key
func NewTestCA(caType types.CertAuthType, clusterName string, privateKeys ...[]byte) *types.CertAuthorityV2 {
	return NewTestCAWithConfig(TestCAConfig{
		Type:        caType,
		ClusterName: clusterName,
		PrivateKeys: privateKeys,
		Clock:       clockwork.NewRealClock(),
	})
}

// NewTestCAWithConfig generates a new certificate authority with the specified
// configuration
// Keep this function in-sync with lib/auth/auth.go:newKeySet().
// TODO(jakule): reuse keystore.KeyStore interface to match newKeySet().
func NewTestCAWithConfig(config TestCAConfig) *types.CertAuthorityV2 {
	var keyPEM []byte
	var key *keys.PrivateKey

	if config.ResourceName == "" {
		config.ResourceName = config.ClusterName
	}
	if config.SubjectOrganization == "" {
		config.SubjectOrganization = config.ClusterName
	}

	switch config.Type {
	case types.DatabaseCA, types.SAMLIDPCA, types.OIDCIdPCA:
		// These CAs only support RSA.
		keyPEM = fixtures.PEMBytes["rsa"]
	case types.DatabaseClientCA:
		// The db client CA also only supports RSA, but some tests rely on it
		// being different than the DB CA.
		keyPEM = fixtures.PEMBytes["rsa-db-client"]
	}
	if len(config.PrivateKeys) > 0 {
		// Allow test to override the private key.
		keyPEM = config.PrivateKeys[0]
	}

	if keyPEM != nil {
		var err error
		key, err = keys.ParsePrivateKey(keyPEM)
		if err != nil {
			panic(err)
		}
	} else {
		// If config.PrivateKeys was not set and this CA does not exclusively
		// support RSA, generate an ECDSA key. Signatures are ~10x faster than
		// RSA and generating a new key is actually faster than parsing a PEM
		// fixture.
		signer, err := cryptosuites.GenerateKeyWithAlgorithm(cryptosuites.ECDSAP256)
		if err != nil {
			panic(err)
		}
		key, err = keys.NewPrivateKey(signer)
		if err != nil {
			panic(err)
		}
		keyPEM = key.PrivateKeyPEM()
	}

	ca := &types.CertAuthorityV2{
		Kind:    types.KindCertAuthority,
		SubKind: string(config.Type),
		Version: types.V2,
		Metadata: types.Metadata{
			Name:      config.ResourceName,
			Namespace: apidefaults.Namespace,
		},
		Spec: types.CertAuthoritySpecV2{
			Type:        config.Type,
			ClusterName: config.ClusterName,
		},
	}

	// Add SSH keys if necessary.
	switch config.Type {
	case types.UserCA, types.HostCA, types.OpenSSHCA:
		ca.Spec.ActiveKeys.SSH = []*types.SSHKeyPair{{
			PrivateKey: keyPEM,
			PublicKey:  key.MarshalSSHPublicKey(),
		}}
	}

	// Add TLS keys if necessary.
	switch config.Type {
	case types.UserCA, types.HostCA, types.DatabaseCA, types.DatabaseClientCA, types.SAMLIDPCA, types.SPIFFECA, types.AWSRACA:
		cert, err := tlsca.GenerateSelfSignedCAWithConfig(tlsca.GenerateCAConfig{
			Signer: key.Signer,
			Entity: pkix.Name{
				CommonName:   config.ClusterName,
				Organization: []string{config.SubjectOrganization},
			},
			TTL:   defaults.CATTL,
			Clock: config.Clock,
		})
		if err != nil {
			panic(err)
		}
		ca.Spec.ActiveKeys.TLS = []*types.TLSKeyPair{{
			Key:  keyPEM,
			Cert: cert,
		}}
	}

	// Add JWT keys if necessary.
	switch config.Type {
	case types.JWTSigner, types.OIDCIdPCA, types.SPIFFECA, types.OktaCA, types.BoundKeypairCA:
		pubKeyPEM, err := keys.MarshalPublicKey(key.Public())
		if err != nil {
			panic(err)
		}
		ca.Spec.ActiveKeys.JWT = []*types.JWTKeyPair{{
			PrivateKey: keyPEM,
			PublicKey:  pubKeyPEM,
		}}
	}

	return ca
}
