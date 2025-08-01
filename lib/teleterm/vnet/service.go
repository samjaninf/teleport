// Teleport
// Copyright (C) 2024 Gravitational, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package vnet

import (
	"context"
	"crypto/tls"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/api/profile"
	"github.com/gravitational/teleport/api/types"
	prehogv1alpha "github.com/gravitational/teleport/gen/proto/go/prehog/v1alpha"
	apiteleterm "github.com/gravitational/teleport/gen/proto/go/teleport/lib/teleterm/v1"
	api "github.com/gravitational/teleport/gen/proto/go/teleport/lib/teleterm/vnet/v1"
	diagv1 "github.com/gravitational/teleport/gen/proto/go/teleport/lib/vnet/diag/v1"
	vnetv1 "github.com/gravitational/teleport/gen/proto/go/teleport/lib/vnet/v1"
	"github.com/gravitational/teleport/lib/client"
	"github.com/gravitational/teleport/lib/teleterm/api/uri"
	"github.com/gravitational/teleport/lib/teleterm/clusteridcache"
	"github.com/gravitational/teleport/lib/teleterm/clusters"
	"github.com/gravitational/teleport/lib/teleterm/daemon"
	logutils "github.com/gravitational/teleport/lib/utils/log"
	"github.com/gravitational/teleport/lib/vnet"
	"github.com/gravitational/teleport/lib/vnet/diag"
)

var log = logutils.NewPackageLogger(teleport.ComponentKey, "term:vnet")

type status int

const (
	statusNotRunning status = iota
	statusRunning
	statusClosed
)

// Service implements gRPC service for VNet.
type Service struct {
	api.UnimplementedVnetServiceServer

	cfg                Config
	mu                 sync.Mutex
	status             status
	usageReporter      usageReporter
	vnetProcess        *vnet.UserProcess
	clusterConfigCache *vnet.ClusterConfigCache
	networkStackInfo   *vnetv1.NetworkStackInfo
}

// New creates an instance of Service.
func New(cfg Config) (*Service, error) {
	if err := cfg.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}

	return &Service{
		cfg:                cfg,
		clusterConfigCache: vnet.NewClusterConfigCache(cfg.Clock),
	}, nil
}

type Config struct {
	// DaemonService is used to get cached clients and for usage reporting. If DaemonService was not
	// one giant blob of methods, Config could accept two separate services instead.
	DaemonService *daemon.Service
	// InsecureSkipVerify signifies whether VNet is going to verify the identity of the proxy service.
	InsecureSkipVerify bool
	// ClusterIDCache is used for usage reporting to read cluster ID that needs to be included with
	// every event.
	ClusterIDCache *clusteridcache.Cache
	// InstallationID is a unique ID of this particular Connect installation, used for usage
	// reporting.
	InstallationID string
	Clock          clockwork.Clock
	profilePath    string
}

// CheckAndSetDefaults checks and sets the defaults
func (c *Config) CheckAndSetDefaults() error {
	if c.DaemonService == nil {
		return trace.BadParameter("missing DaemonService")
	}

	if c.ClusterIDCache == nil {
		return trace.BadParameter("missing ClusterIDCache")
	}

	if c.InstallationID == "" {
		return trace.BadParameter("missing InstallationID")
	}

	if c.Clock == nil {
		c.Clock = clockwork.NewRealClock()
	}

	if c.profilePath == "" {
		c.profilePath = profile.FullProfilePath(os.Getenv(types.HomeEnvVar))
	}

	return nil
}

func (s *Service) Start(ctx context.Context, req *api.StartRequest) (*api.StartResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status == statusClosed {
		return nil, trace.CompareFailed("VNet service has been closed")
	}

	if s.status == statusRunning {
		return nil, trace.AlreadyExists("VNet is already running")
	}

	clientApplication := &clientApplication{
		daemonService:      s.cfg.DaemonService,
		insecureSkipVerify: s.cfg.InsecureSkipVerify,
		usageReporter:      &disabledTelemetryUsageReporter{},
	}

	// Generally, the usage reporting setting cannot be changed without restarting the app, so
	// technically this information could have been passed through argv to tsh daemon.
	// However, there is one exception: during the first launch of the app, the user is asked if they
	// want to enable telemetry. Agreeing to that changes the setting without restarting the app.
	// As such, this service needs to ask for this setting on every launch.
	isUsageReportingEnabled, err := s.isUsageReportingEnabled(ctx)
	if err != nil {
		return nil, trace.Wrap(err, "getting usage reporting settings")
	}

	if isUsageReportingEnabled {
		usageReporter, err := newDaemonUsageReporter(daemonUsageReporterConfig{
			ClientCache:    s.cfg.DaemonService,
			EventConsumer:  s.cfg.DaemonService,
			ClusterIDCache: s.cfg.ClusterIDCache,
			InstallationID: s.cfg.InstallationID,
		})
		if err != nil {
			return nil, trace.Wrap(err)
		}
		defer func() {
			if s.status != statusRunning {
				usageReporter.Stop()
			}
		}()
		clientApplication.usageReporter = usageReporter
	}

	vnetProcess, err := vnet.RunUserProcess(ctx, clientApplication)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	go func() {
		err := vnetProcess.Wait()
		if err != nil && !errors.Is(err, context.Canceled) {
			log.ErrorContext(ctx, "VNet closed with an error", "error", err)
		} else {
			log.DebugContext(ctx, "VNet closed")
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		// Handle unexpected shutdown.
		// If vnetProcess.Wait has returned but status is stil "running", then it means that VNet
		// unexpectedly shut down rather than stopped through the Stop RPC.
		if s.status == statusRunning {
			s.status = statusNotRunning

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			reportErr := s.reportUnexpectedShutdown(ctx, err)
			if reportErr != nil {
				log.ErrorContext(ctx, "Could not notify the Electron app about unexpected VNet shutdown",
					"shutdown_error", err, "notify_error", reportErr)
			}
		}
	}()

	s.vnetProcess = vnetProcess
	s.networkStackInfo = vnetProcess.NetworkStackInfo()
	s.usageReporter = clientApplication.usageReporter
	s.status = statusRunning
	return &api.StartResponse{}, nil
}

// Stop stops VNet and cleans up used resources. Blocks until VNet stops or ctx is canceled.
func (s *Service) Stop(ctx context.Context, req *api.StopRequest) (*api.StopResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.stopLocked()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &api.StopResponse{}, nil
}

// GetServiceInfo returns info about the running VNet service.
func (s *Service) GetServiceInfo(ctx context.Context, _ *api.GetServiceInfoRequest) (*api.GetServiceInfoResponse, error) {
	// Acquire the lock just to check the status of the service. We don't want the actual process of
	// listing DNS zones to block the user from performing other operations.
	s.mu.Lock()
	if s.status != statusRunning {
		s.mu.Unlock()
		return nil, trace.CompareFailed("VNet is not running")
	}
	unifiedClusterConfigProvider := s.vnetProcess.GetUnifiedClusterConfigProvider()
	s.mu.Unlock()

	unifiedClusterConfig, err := unifiedClusterConfigProvider.GetUnifiedClusterConfig(ctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	sshConfigChecker, err := diag.NewSSHConfigChecker(s.cfg.profilePath)
	if err != nil {
		return nil, trace.Wrap(err, "building SSH config checker")
	}
	_, sshConfigured, err := sshConfigChecker.OpenSSHConfigIncludesVNetSSHConfig()
	if err != nil && !trace.IsNotFound(err) {
		return nil, trace.Wrap(err, "checking SSH configuration")
	}

	return &api.GetServiceInfoResponse{
		AppDnsZones:       unifiedClusterConfig.AppDNSZones(),
		Clusters:          unifiedClusterConfig.ClusterNames,
		SshConfigured:     sshConfigured,
		VnetSshConfigPath: sshConfigChecker.VNetSSHConfigPath,
	}, nil
}

// RunDiagnostics runs a set of heuristics to determine if VNet actually works
// on the device. It requires VNet to be started.
func (s *Service) RunDiagnostics(ctx context.Context, req *api.RunDiagnosticsRequest) (*api.RunDiagnosticsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status != statusRunning {
		return nil, trace.CompareFailed("VNet is not running")
	}

	if s.networkStackInfo.InterfaceName == "" {
		return nil, trace.BadParameter("no interface name, this is a bug")
	}

	if s.networkStackInfo.Ipv6Prefix == "" {
		return nil, trace.BadParameter("no IPv6 prefix, this is a bug")
	}

	nsa := &diagv1.NetworkStackAttempt{}
	if ns, err := s.getNetworkStack(ctx); err != nil {
		nsa.Status = diagv1.CheckAttemptStatus_CHECK_ATTEMPT_STATUS_ERROR
		nsa.Error = err.Error()
	} else {
		nsa.Status = diagv1.CheckAttemptStatus_CHECK_ATTEMPT_STATUS_OK
		nsa.NetworkStack = ns
	}

	diagChecks, err := s.platformDiagChecks(ctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	report, err := diag.GenerateReport(ctx, diag.ReportPrerequisites{
		Clock:               s.cfg.Clock,
		NetworkStackAttempt: nsa,
		DiagChecks:          diagChecks,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &api.RunDiagnosticsResponse{
		Report: report,
	}, nil
}

func (s *Service) getNetworkStack(ctx context.Context) (*diagv1.NetworkStack, error) {
	unifiedClusterConfig, err := s.vnetProcess.GetUnifiedClusterConfigProvider().GetUnifiedClusterConfig(ctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &diagv1.NetworkStack{
		InterfaceName:  s.networkStackInfo.InterfaceName,
		Ipv6Prefix:     s.networkStackInfo.Ipv6Prefix,
		Ipv4CidrRanges: unifiedClusterConfig.IPv4CidrRanges,
		DnsZones:       unifiedClusterConfig.AllDNSZones(),
	}, nil
}

// AutoConfigureSSH automatically configures OpenSSH-compatible clients for
// connections to Teleport SSH servers through VNet.
func (s *Service) AutoConfigureSSH(ctx context.Context, _ *api.AutoConfigureSSHRequest) (*api.AutoConfigureSSHResponse, error) {
	err := vnet.AutoConfigureOpenSSH(ctx, s.cfg.profilePath)
	return nil, trace.Wrap(err)
}

func (s *Service) stopLocked() error {
	if s.status == statusClosed {
		return trace.CompareFailed("VNet service has been closed")
	}

	if s.status == statusNotRunning {
		return nil
	}

	s.vnetProcess.Close()
	err := s.vnetProcess.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return trace.Wrap(err)
	}
	s.usageReporter.Stop()

	s.status = statusNotRunning
	return nil
}

// Close stops VNet service and prevents it from being started again. Blocks until VNet stops.
// Intended for cleanup code when tsh daemon gets terminated.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.stopLocked()
	if err != nil {
		return trace.Wrap(err)
	}

	s.status = statusClosed
	return nil
}

func (s *Service) isUsageReportingEnabled(ctx context.Context) (bool, error) {
	tshdEventsClient, err := s.cfg.DaemonService.TshdEventsClient(ctx)
	if err != nil {
		return false, trace.Wrap(err)
	}

	resp, err := tshdEventsClient.GetUsageReportingSettings(ctx, &apiteleterm.GetUsageReportingSettingsRequest{})
	if err != nil {
		return false, trace.Wrap(err)
	}

	return resp.UsageReportingSettings.Enabled, nil
}

func (s *Service) reportUnexpectedShutdown(ctx context.Context, shutdownErr error) error {
	tshdEventsClient, err := s.cfg.DaemonService.TshdEventsClient(ctx)
	if err != nil {
		return trace.Wrap(err, "obtaining tshd events client")
	}

	var shutdownErrorMsg string
	if shutdownErr != nil {
		shutdownErrorMsg = shutdownErr.Error()
	}

	_, err = tshdEventsClient.ReportUnexpectedVnetShutdown(ctx, &apiteleterm.ReportUnexpectedVnetShutdownRequest{
		Error: shutdownErrorMsg,
	})
	return trace.Wrap(err, "sending shutdown report")
}

type clientApplication struct {
	daemonService      *daemon.Service
	usageReporter      usageReporter
	insecureSkipVerify bool
}

func (p *clientApplication) ListProfiles() ([]string, error) {
	profiles, err := p.daemonService.ListProfileNames()
	return profiles, trace.Wrap(err)
}

func (p *clientApplication) GetCachedClient(ctx context.Context, profileName, leafClusterName string) (vnet.ClusterClient, error) {
	return p.getCachedClient(ctx, profileName, leafClusterName)
}

func (p *clientApplication) getCachedClient(ctx context.Context, profileName, leafClusterName string) (*client.ClusterClient, error) {
	uri := uri.NewClusterURI(profileName).AppendLeafCluster(leafClusterName)
	client, err := p.daemonService.GetCachedClient(ctx, uri)
	return client, trace.Wrap(err)
}

func (p *clientApplication) ReissueAppCert(ctx context.Context, appInfo *vnetv1.AppInfo, targetPort uint16) (tls.Certificate, error) {
	appKey := appInfo.GetAppKey()
	clusterURI := uri.NewClusterURI(appKey.GetProfile()).AppendLeafCluster(appKey.GetLeafCluster())
	appURI := clusterURI.AppendApp(appKey.GetName())

	routeToApp := vnet.RouteToApp(appInfo, targetPort)
	apiteletermRouteToApp := apiteleterm.RouteToApp{
		Name:        routeToApp.Name,
		PublicAddr:  routeToApp.PublicAddr,
		ClusterName: routeToApp.ClusterName,
		Uri:         routeToApp.URI,
		TargetPort:  routeToApp.TargetPort,
	}

	reloginReq := &apiteleterm.ReloginRequest{
		RootClusterUri: clusterURI.GetRootClusterURI().String(),
		Reason: &apiteleterm.ReloginRequest_VnetCertExpired{
			VnetCertExpired: &apiteleterm.VnetCertExpired{
				TargetUri:  appURI.String(),
				RouteToApp: &apiteletermRouteToApp,
			},
		},
	}

	var cert tls.Certificate

	reissueCert := func() error {
		cluster, _, err := p.daemonService.ResolveClusterURI(clusterURI)
		if err != nil {
			return trace.Wrap(err)
		}

		client, err := p.daemonService.GetCachedClient(ctx, clusterURI)
		if err != nil {
			return trace.Wrap(err)
		}

		cert, err = cluster.ReissueAppCert(ctx, client, *routeToApp)
		return trace.Wrap(err)
	}

	if err := p.daemonService.RetryWithRelogin(ctx, reloginReq, reissueCert); err != nil {
		notifyErr := p.daemonService.NotifyApp(ctx, &apiteleterm.SendNotificationRequest{
			Subject: &apiteleterm.SendNotificationRequest_CannotProxyVnetConnection{
				CannotProxyVnetConnection: &apiteleterm.CannotProxyVnetConnection{
					TargetUri:  appURI.String(),
					RouteToApp: &apiteletermRouteToApp,
					Reason: &apiteleterm.CannotProxyVnetConnection_CertReissueError{
						CertReissueError: &apiteleterm.CertReissueError{
							Error: err.Error(),
						},
					},
				},
			},
		})
		if notifyErr != nil {
			log.ErrorContext(ctx, "Failed to send a notification for an error encountered during VNet cert reissue",
				"cert_reissue_error", err, "notify_error", notifyErr)
		}

		return tls.Certificate{}, trace.Wrap(err)
	}

	return cert, nil
}

// UserTLSCert returns the user TLS certificate for the given profile.
func (p *clientApplication) UserTLSCert(ctx context.Context, profileName string) (tls.Certificate, error) {
	// We don't have easy access to the user TLS cert from here, the only way
	// I've found is to reach through the ProxyClient as this does below.
	clusterClient, err := p.getCachedClient(ctx, profileName, "")
	if err != nil {
		return tls.Certificate{}, trace.Wrap(err)
	}
	clientConfig, err := clusterClient.ProxyClient.ClientConfig(ctx, "")
	if err != nil {
		return tls.Certificate{}, trace.Wrap(err, "getting user client config")
	}
	if len(clientConfig.Credentials) < 1 {
		return tls.Certificate{}, trace.Errorf("user client config has no credentials")
	}
	cred := clientConfig.Credentials[0]
	tlsConfig, err := cred.TLSConfig()
	if err != nil {
		return tls.Certificate{}, trace.Wrap(err, "getting user TLS config")
	}
	switch {
	case len(tlsConfig.Certificates) > 0:
		return tlsConfig.Certificates[0], nil
	case tlsConfig.GetClientCertificate != nil:
		// This is the actual path we currently take at the time of writing,
		// api/client.configureTLS always sets tlsConfig.GetClientCertificate
		// and unsets tlsConfig.Certificates.
		tlsCert, err := tlsConfig.GetClientCertificate(nil)
		if err != nil {
			return tls.Certificate{}, trace.Wrap(err, "getting client TLS certificate")
		}
		return *tlsCert, nil
	default:
		return tls.Certificate{}, trace.Errorf("user TLS config has no certificates")
	}
}

// GetDialOptions returns ALPN dial options for the profile.
func (p *clientApplication) GetDialOptions(ctx context.Context, profileName string) (*vnetv1.DialOptions, error) {
	cluster, tc, err := p.daemonService.ResolveClusterURI(uri.NewClusterURI(profileName))
	if err != nil {
		return nil, trace.Wrap(err, "resolving cluster by URI")
	}

	dialOpts := &vnetv1.DialOptions{
		WebProxyAddr:            cluster.GetProxyHost(),
		AlpnConnUpgradeRequired: tc.TLSRoutingConnUpgradeRequired,
		InsecureSkipVerify:      p.insecureSkipVerify,
	}
	dialOpts.RootClusterCaCertPool, err = tc.RootClusterCACertPoolPEM(ctx)
	if err != nil {
		return nil, trace.Wrap(err, "loading root cluster CA cert pool")
	}
	return dialOpts, nil
}

// OnNewSSHSession submits a usage event for a new SSH session.
func (p *clientApplication) OnNewSSHSession(ctx context.Context, profileName, targetClusterName string) {
	// Enqueue the event from a separate goroutine since we don't care about errors anyway and we also
	// don't want to slow down VNet connections.
	go func() {
		// Not passing ctx to ReportSSHSession since ctx is tied to the
		// lifetime of a short-lived API call, inheriting the context could
		// interrupt reporting.
		if err := p.usageReporter.ReportSSHSession(profileName, targetClusterName); err != nil {
			log.ErrorContext(ctx, "Failed to submit SSH usage event")
		}
	}()
}

// OnNewAppConnection submits an app usage event once per clientApplication lifetime.
// That is, if a user makes multiple connections to a single app, OnNewAppConnection submits a single
// event. This is to mimic how Connect submits events for its app gateways. This lets us compare
// popularity of VNet and app gateways.
func (p *clientApplication) OnNewAppConnection(ctx context.Context, appKey *vnetv1.AppKey) error {
	// Enqueue the event from a separate goroutine since we don't care about errors anyway and we also
	// don't want to slow down VNet connections.
	go func() {
		uri := uri.NewClusterURI(appKey.GetProfile()).AppendLeafCluster(appKey.GetLeafCluster()).AppendApp(appKey.GetName())

		// Not passing ctx to ReportApp since ctx is tied to the lifetime of the connection.
		// If it's a short-lived connection, inheriting its context would interrupt reporting.
		if err := p.usageReporter.ReportApp(uri); err != nil {
			log.ErrorContext(ctx, "Failed to submit app usage event", "app", uri, "error", err)
		}
	}()

	return nil
}

// OnInvalidLocalPort gets called before VNet refuses to handle a connection to a multi-port TCP app
// because the provided port does not match any of the TCP ports in the app spec.
func (p *clientApplication) OnInvalidLocalPort(ctx context.Context, appInfo *vnetv1.AppInfo, targetPort uint16) {
	// If something is wrong with the Electron app to the point that it stopped accepting RPCs, return
	// quickly rather than being blocked on sending a notification.
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	appKey := appInfo.GetAppKey()
	appURI := uri.NewClusterURI(appKey.GetProfile()).
		AppendLeafCluster(appKey.GetLeafCluster()).
		AppendApp(appKey.GetName())
	routeToApp := vnet.RouteToApp(appInfo, targetPort)
	apiteletermRouteToApp := apiteleterm.RouteToApp{
		Name:        routeToApp.Name,
		PublicAddr:  routeToApp.PublicAddr,
		ClusterName: routeToApp.ClusterName,
		Uri:         routeToApp.URI,
		TargetPort:  routeToApp.TargetPort,
	}

	invalidLocalPort := &apiteleterm.InvalidLocalPort{}
	// Send ports only if there's less than 10 ranges. A bigger number would be difficult to show in
	// the UI.
	tcpPorts := appInfo.GetApp().GetTCPPorts()
	if len(tcpPorts) <= 10 {
		apiTCPPorts := make([]*apiteleterm.PortRange, 0, len(tcpPorts))
		for _, portRange := range tcpPorts {
			apiTCPPorts = append(apiTCPPorts, &apiteleterm.PortRange{Port: portRange.Port, EndPort: portRange.EndPort})
		}
		invalidLocalPort.TcpPorts = apiTCPPorts
	}

	err := p.daemonService.NotifyApp(ctx, &apiteleterm.SendNotificationRequest{
		Subject: &apiteleterm.SendNotificationRequest_CannotProxyVnetConnection{
			CannotProxyVnetConnection: &apiteleterm.CannotProxyVnetConnection{
				TargetUri:  appURI.String(),
				RouteToApp: &apiteletermRouteToApp,
				Reason: &apiteleterm.CannotProxyVnetConnection_InvalidLocalPort{
					InvalidLocalPort: invalidLocalPort,
				},
			},
		},
	})
	if err != nil {
		log.ErrorContext(ctx, "Could not notify the Electron app about invalid local port",
			"notify_error", err,
			"profile_name", appKey.GetProfile(),
			"leaf_cluster_name", appKey.GetLeafCluster(),
			"route_to_app", routeToApp)
	}
}

type usageReporter interface {
	ReportApp(uri.ResourceURI) error
	ReportSSHSession(profileName, rootClusterName string) error
	Stop()
}

type daemonUsageReporter struct {
	cfg daemonUsageReporterConfig
	// reportedApps contains a set of URIs for apps which usage has been already reported.
	// App gateways (local proxies) in Connect report a single event per gateway created per app. VNet
	// needs to replicate this behavior, hence why it keeps track of reported apps to report only one
	// event per app per VNet's lifespan.
	reportedApps map[string]struct{}
	// mu protects access to reportedApps.
	mu sync.Mutex
	// close is used to abort a ReportApp call that's currently in flight.
	close chan struct{}
	// closed signals that usageReporter has been stopped and no more events should be reported.
	closed atomic.Bool
}

type clientCache interface {
	GetCachedClient(context.Context, uri.ResourceURI) (*client.ClusterClient, error)
	ResolveClusterURI(uri uri.ResourceURI) (*clusters.Cluster, *client.TeleportClient, error)
}

type eventConsumer interface {
	ReportUsageEvent(*apiteleterm.ReportUsageEventRequest) error
}

type daemonUsageReporterConfig struct {
	ClientCache   clientCache
	EventConsumer eventConsumer
	// clusterIDCache stores cluster ID that needs to be included with each usage event. It's updated
	// outside of usageReporter – the middleware merely reads data from it. If the cache does not
	// contain the given cluster ID, usageReporter drops the event.
	ClusterIDCache *clusteridcache.Cache
	InstallationID string
}

func (c *daemonUsageReporterConfig) CheckAndSetDefaults() error {
	if c.ClientCache == nil {
		return trace.BadParameter("missing ClientCache")
	}

	if c.EventConsumer == nil {
		return trace.BadParameter("missing EventConsumer")
	}

	if c.ClusterIDCache == nil {
		return trace.BadParameter("missing ClusterIDCache")
	}

	if c.InstallationID == "" {
		return trace.BadParameter("missing InstallationID")
	}

	return nil
}

func newDaemonUsageReporter(cfg daemonUsageReporterConfig) (*daemonUsageReporter, error) {
	if err := cfg.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}

	return &daemonUsageReporter{
		cfg:          cfg,
		reportedApps: make(map[string]struct{}),
		close:        make(chan struct{}),
	}, nil
}

// ReportSSHSession adds an event for a new SSH session to the events queue.
// It reports a new event for each new SSH session, in contrast to ReportApp
// which only reports each unique app once, to align with how Connect reports
// usage events for SSH sessions.
func (r *daemonUsageReporter) ReportSSHSession(profileName, rootClusterName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed.Load() {
		return trace.CompareFailed("usage reporter has been stopped")
	}

	rootClusterURI := uri.NewClusterURI(profileName)
	_, tc, err := r.cfg.ClientCache.ResolveClusterURI(rootClusterURI)
	if err != nil {
		return trace.Wrap(err)
	}

	clusterID, ok := r.cfg.ClusterIDCache.Load(rootClusterURI)
	if !ok {
		return trace.NotFound("cluster ID for %q not found", rootClusterURI)
	}

	log.DebugContext(context.Background(), "Reporting SSH usage event", "profile", profileName, "root_cluster", rootClusterName)
	if err := r.cfg.EventConsumer.ReportUsageEvent(&apiteleterm.ReportUsageEventRequest{
		AuthClusterId: clusterID,
		PrehogReq: &prehogv1alpha.SubmitConnectEventRequest{
			DistinctId: r.cfg.InstallationID,
			Timestamp:  timestamppb.Now(),
			Event: &prehogv1alpha.SubmitConnectEventRequest_ProtocolUse{
				ProtocolUse: &prehogv1alpha.ConnectProtocolUseEvent{
					ClusterName:   rootClusterName,
					UserName:      tc.Username,
					Protocol:      "ssh",
					Origin:        "vnet",
					AccessThrough: "vnet",
				},
			},
		},
	}); err != nil {
		return trace.Wrap(err, "adding SSH usage event to queue")
	}
	return nil
}

// ReportApp adds an event related to the given app to the events queue, if the app wasn't reported
// already. Only one invocation of ReportApp can be in flight at a time.
func (r *daemonUsageReporter) ReportApp(appURI uri.ResourceURI) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed.Load() {
		return trace.CompareFailed("usage reporter has been stopped")
	}

	if _, hasAppBeenReported := r.reportedApps[appURI.String()]; hasAppBeenReported {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case <-r.close:
			cancel()
		case <-ctx.Done():
		}
	}()

	rootClusterURI := appURI.GetRootClusterURI()
	client, err := r.cfg.ClientCache.GetCachedClient(ctx, appURI)
	if err != nil {
		return trace.Wrap(err)
	}
	rootClusterName := client.RootClusterName()
	_, tc, err := r.cfg.ClientCache.ResolveClusterURI(appURI)
	if err != nil {
		return trace.Wrap(err)
	}

	clusterID, ok := r.cfg.ClusterIDCache.Load(rootClusterURI)
	if !ok {
		return trace.NotFound("cluster ID for %q not found", rootClusterURI)
	}

	log.DebugContext(ctx, "Reporting app usage event", "app", appURI.String())

	if err := r.cfg.EventConsumer.ReportUsageEvent(&apiteleterm.ReportUsageEventRequest{
		AuthClusterId: clusterID,
		PrehogReq: &prehogv1alpha.SubmitConnectEventRequest{
			DistinctId: r.cfg.InstallationID,
			Timestamp:  timestamppb.Now(),
			Event: &prehogv1alpha.SubmitConnectEventRequest_ProtocolUse{
				ProtocolUse: &prehogv1alpha.ConnectProtocolUseEvent{
					ClusterName:   rootClusterName,
					UserName:      tc.Username,
					Protocol:      "app",
					Origin:        "vnet",
					AccessThrough: "vnet",
				},
			},
		},
	}); err != nil {
		return trace.Wrap(err, "adding app usage event to queue")
	}

	r.reportedApps[appURI.String()] = struct{}{}

	return nil
}

// Stop aborts the reporting of an event that's currently in progress and prevents further events
// from being reported. It blocks until the current ReportApp call aborts.
func (r *daemonUsageReporter) Stop() {
	if r.closed.Load() {
		return
	}

	// Prevent new calls to ReportApp from being made.
	r.closed.Store(true)
	// Abort context of the ReportApp call currently in flight.
	close(r.close)
	// Block until the current ReportApp call aborts.
	r.mu.Lock()
	defer r.mu.Unlock()
}

type disabledTelemetryUsageReporter struct{}

func (r *disabledTelemetryUsageReporter) ReportApp(appURI uri.ResourceURI) error {
	log.DebugContext(context.Background(), "Skipping app usage event, usage reporting is turned off", "app", appURI.String())
	return nil
}

func (r *disabledTelemetryUsageReporter) ReportSSHSession(profileName, rootClusterName string) error {
	log.DebugContext(context.Background(), "Skipping SSH usage event, usage reporting is turned off", "profile", profileName, "root_cluster", rootClusterName)
	return nil
}

func (r *disabledTelemetryUsageReporter) Stop() {}
