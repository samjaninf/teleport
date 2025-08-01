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

package inventory

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/api/client"
	"github.com/gravitational/teleport/api/client/proto"
	presencev1 "github.com/gravitational/teleport/api/gen/proto/go/teleport/presence/v1"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/inventory/metadata"
	usagereporter "github.com/gravitational/teleport/lib/usagereporter/teleport"
	"github.com/gravitational/teleport/lib/utils"
	"github.com/gravitational/teleport/lib/utils/log/logtest"
)

func TestMain(m *testing.M) {
	logtest.InitLogger(testing.Verbose)
	os.Exit(m.Run())
}

type fakeAuth struct {
	mu             sync.Mutex
	failUpserts    int
	failKeepAlives int
	failDeletes    int

	upserts    int
	keepalives int
	deletes    int
	err        error

	expectAddr      string
	unexpectedAddrs int

	failUpsertInstance int

	lastInstance    types.Instance
	lastRawInstance []byte

	lastServerExpiry time.Time
}

func (a *fakeAuth) getLastServerExpiry() time.Time {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.lastServerExpiry
}

func (a *fakeAuth) UpsertNode(_ context.Context, server types.Server) (*types.KeepAlive, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.upserts++
	if a.expectAddr != "" {
		if server.GetAddr() != a.expectAddr {
			a.unexpectedAddrs++
		}
	}
	if a.failUpserts > 0 {
		a.failUpserts--
		return nil, trace.Errorf("upsert failed as test condition")
	}
	a.lastServerExpiry = server.Expiry()
	return &types.KeepAlive{}, a.err
}

func (a *fakeAuth) UpsertApplicationServer(_ context.Context, server types.AppServer) (*types.KeepAlive, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.upserts++

	if a.failUpserts > 0 {
		a.failUpserts--
		return nil, trace.Errorf("upsert failed as test condition")
	}
	a.lastServerExpiry = server.Expiry()
	return &types.KeepAlive{}, a.err
}

func (a *fakeAuth) DeleteApplicationServer(ctx context.Context, namespace, hostID, name string) error {
	return nil
}

func (a *fakeAuth) UpsertDatabaseServer(_ context.Context, server types.DatabaseServer) (*types.KeepAlive, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.upserts++

	if a.failUpserts > 0 {
		a.failUpserts--
		return nil, trace.Errorf("upsert failed as test condition")
	}
	a.lastServerExpiry = server.Expiry()
	return &types.KeepAlive{}, a.err
}

func (a *fakeAuth) DeleteDatabaseServer(ctx context.Context, namespace, hostID, name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.deletes++

	if a.failDeletes > 0 {
		a.failDeletes--
		return trace.Errorf("delete failed as test condition")
	}
	return nil
}

func (a *fakeAuth) UpsertKubernetesServer(_ context.Context, server types.KubeServer) (*types.KeepAlive, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.upserts++

	if a.failUpserts > 0 {
		a.failUpserts--
		return nil, trace.Errorf("upsert failed as test condition")
	}
	a.lastServerExpiry = server.Expiry()
	return &types.KeepAlive{}, a.err
}

func (a *fakeAuth) DeleteKubernetesServer(ctx context.Context, hostID, name string) error {
	return nil
}

func (a *fakeAuth) KeepAliveServer(_ context.Context, ka types.KeepAlive) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.keepalives++
	if a.failKeepAlives > 0 {
		a.failKeepAlives--
		return trace.Errorf("keepalive failed as test condition")
	}
	a.lastServerExpiry = ka.Expires
	return a.err
}

// UpsertRelayServer implements [Auth].
func (a *fakeAuth) UpsertRelayServer(ctx context.Context, relayServer *presencev1.RelayServer) (*presencev1.RelayServer, error) {
	panic("unimplemented")
}

// DeleteRelayServer implements [Auth].
func (a *fakeAuth) DeleteRelayServer(ctx context.Context, name string) error {
	panic("unimplemented")
}

func (a *fakeAuth) UpsertInstance(ctx context.Context, instance types.Instance) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.failUpsertInstance > 0 {
		a.failUpsertInstance--
		return trace.Errorf("upsert instance failed as test condition")
	}

	a.lastInstance = instance.Clone()
	var err error
	a.lastRawInstance, err = utils.FastMarshal(instance)
	if err != nil {
		panic("fastmarshal of instance should be infallible")
	}

	return nil
}

// TestSSHServerBasics verifies basic expected behaviors for a single control stream heartbeating
// an ssh service.
func TestSSHServerBasics(t *testing.T) {
	const serverID = "test-server"
	const zeroAddr = "0.0.0.0:123"
	const peerAddr = "1.2.3.4:456"
	const wantAddr = "1.2.3.4:123"

	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan testEvent, 1024)

	auth := &fakeAuth{
		expectAddr: wantAddr,
	}

	rc := &resourceCounter{}
	controller := NewController(
		auth,
		usagereporter.DiscardUsageReporter{},
		withServerKeepAlive(time.Millisecond*200),
		withTestEventsChannel(events),
		WithOnConnect(rc.onConnect),
		WithOnDisconnect(rc.onDisconnect),
	)
	defer controller.Close()

	// set up fake in-memory control stream
	upstream, downstream := client.InventoryControlStreamPipe(client.ICSPipePeerAddr(peerAddr))
	t.Cleanup(func() {
		controller.Close()
		downstream.Close()
		upstream.Close()
	})

	// launch goroutine to respond to ping requests
	go func() {
		for {
			select {
			case msg := <-downstream.Recv():
				downstream.Send(ctx, &proto.UpstreamInventoryPong{
					ID: msg.(*proto.DownstreamInventoryPing).GetID(),
				})
			case <-downstream.Done():
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	controller.RegisterControlStream(upstream, &proto.UpstreamInventoryHello{
		ServerID: serverID,
		Version:  teleport.Version,
		Services: types.SystemRoles{types.RoleNode}.StringSlice(),
	})

	// verify that control stream handle is now accessible
	handle, ok := controller.GetControlStream(serverID)
	require.True(t, ok)

	// verify that hb counter has been incremented
	require.Equal(t, int64(1), controller.instanceHBVariableDuration.Count())

	// send a fake ssh server heartbeat
	err := downstream.Send(ctx, &proto.InventoryHeartbeat{
		SSHServer: &types.ServerV2{
			Metadata: types.Metadata{
				Name: serverID,
			},
			Spec: types.ServerSpecV2{
				Addr: zeroAddr,
			},
		},
	})
	require.NoError(t, err)

	// verify that heartbeat creates both an upsert and a keepalive
	awaitEvents(t, events,
		expect(sshUpsertOk, sshKeepAliveOk),
		deny(sshUpsertErr, sshKeepAliveErr, handlerClose),
	)

	// we will check that the expiration time will grow after keepalives and new
	// server announces
	expiry := auth.getLastServerExpiry()

	// set up to induce some failures, but not enough to cause the control
	// stream to be closed.
	auth.mu.Lock()
	auth.failUpserts = 2
	auth.mu.Unlock()

	// keepalive should fail twice, but since the upsert is already known
	// to have succeeded, we should not see an upsert failure yet.
	awaitEvents(t, events,
		expect(sshKeepAliveErr, sshKeepAliveErr, sshKeepAliveOk),
		deny(sshUpsertErr, handlerClose),
	)

	err = downstream.Send(ctx, &proto.InventoryHeartbeat{
		SSHServer: &types.ServerV2{
			Metadata: types.Metadata{
				Name: serverID,
			},
			Spec: types.ServerSpecV2{
				Addr: zeroAddr,
			},
		},
	})
	require.NoError(t, err)

	// this explicit upsert will not happen since the server is the same, but
	// keepalives should work
	awaitEvents(t, events,
		expect(sshKeepAliveOk),
		deny(sshKeepAliveErr, sshUpsertErr, sshUpsertRetryOk, handlerClose),
	)

	oldExpiry, expiry := expiry, auth.getLastServerExpiry()
	require.Greater(t, expiry, oldExpiry)

	err = downstream.Send(ctx, &proto.InventoryHeartbeat{
		SSHServer: &types.ServerV2{
			Metadata: types.Metadata{
				Name: serverID,
				Labels: map[string]string{
					"changed": "changed",
				},
			},
			Spec: types.ServerSpecV2{
				Addr: zeroAddr,
			},
		},
	})
	require.NoError(t, err)

	auth.mu.Lock()
	auth.failUpserts = 1
	auth.mu.Unlock()

	// we should now see an upsert failure, but no additional
	// keepalive failures, and the upsert should succeed on retry.
	awaitEvents(t, events,
		expect(sshKeepAliveOk, sshUpsertErr, sshUpsertRetryOk),
		deny(sshKeepAliveErr, handlerClose),
	)

	oldExpiry, expiry = expiry, auth.getLastServerExpiry()
	require.Greater(t, expiry, oldExpiry)

	// limit time of ping call
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// execute ping
	_, err = handle.Ping(pingCtx, 1)
	require.NoError(t, err)

	// set up to induce enough consecutive errors to cause stream closure
	auth.mu.Lock()
	auth.failUpserts = 5
	auth.mu.Unlock()

	err = downstream.Send(ctx, &proto.InventoryHeartbeat{
		SSHServer: &types.ServerV2{
			Metadata: types.Metadata{
				Name: serverID,
			},
			Spec: types.ServerSpecV2{
				Addr: zeroAddr,
			},
		},
	})
	require.NoError(t, err)

	// both the initial upsert and the retry should fail, then the handle should
	// close.
	awaitEvents(t, events,
		expect(sshUpsertErr, sshUpsertRetryErr, handlerClose),
		deny(sshUpsertOk),
	)

	// verify that closure propagates to server and client side interfaces
	closeTimeout := time.After(time.Second * 10)
	select {
	case <-handle.Done():
	case <-closeTimeout:
		t.Fatal("timeout waiting for handle closure")
	}
	select {
	case <-downstream.Done():
	case <-closeTimeout:
		t.Fatal("timeout waiting for handle closure")
	}

	// verify that hb counter has been decremented (counter is decremented concurrently, but
	// always *before* closure is propagated to downstream handle, hence being safe to load
	// here).
	require.Equal(t, int64(0), controller.instanceHBVariableDuration.Count())

	// verify that metrics have been updated correctly
	require.Zero(t, 0, rc.count())

	// verify that the peer address of the control stream was used to override
	// zero-value IPs for heartbeats.
	auth.mu.Lock()
	unexpectedAddrs := auth.unexpectedAddrs
	auth.mu.Unlock()
	require.Zero(t, unexpectedAddrs)
}

// TestAppServerBasics verifies basic expected behaviors for a single control stream heartbeating
// an app service.
func TestAppServerBasics(t *testing.T) {
	const serverID = "test-server"
	const appCount = 3

	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan testEvent, 1024)

	auth := &fakeAuth{}

	rc := &resourceCounter{}
	controller := NewController(
		auth,
		usagereporter.DiscardUsageReporter{},
		withServerKeepAlive(time.Millisecond*200),
		withTestEventsChannel(events),
		WithOnConnect(rc.onConnect),
		WithOnDisconnect(rc.onDisconnect),
	)
	defer controller.Close()

	// set up fake in-memory control stream
	upstream, downstream := client.InventoryControlStreamPipe()
	t.Cleanup(func() {
		controller.Close()
		upstream.Close()
		downstream.Close()
	})

	// launch goroutine to respond to ping requests
	go func() {
		for {
			select {
			case msg := <-downstream.Recv():
				downstream.Send(ctx, &proto.UpstreamInventoryPong{
					ID: msg.(*proto.DownstreamInventoryPing).GetID(),
				})
			case <-downstream.Done():
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	controller.RegisterControlStream(upstream, &proto.UpstreamInventoryHello{
		ServerID: serverID,
		Version:  teleport.Version,
		Services: types.SystemRoles{types.RoleApp}.StringSlice(),
	})

	// verify that control stream handle is now accessible
	handle, ok := controller.GetControlStream(serverID)
	require.True(t, ok)

	// verify that hb counter has been incremented
	require.Equal(t, int64(1), controller.instanceHBVariableDuration.Count())

	// send a fake app server heartbeat
	for i := range appCount {
		err := downstream.Send(ctx, &proto.InventoryHeartbeat{
			AppServer: &types.AppServerV3{
				Metadata: types.Metadata{
					Name: serverID,
				},
				Spec: types.AppServerSpecV3{
					HostID: serverID,
					App: &types.AppV3{
						Kind:    types.KindApp,
						Version: types.V3,
						Metadata: types.Metadata{
							Name: fmt.Sprintf("app-%d", i),
						},
						Spec: types.AppSpecV3{},
					},
				},
			},
		})
		require.NoError(t, err)
	}

	// verify that heartbeat creates both an upsert and a keepalive
	awaitEvents(t, events,
		expect(appUpsertOk, appKeepAliveOk, appUpsertOk, appKeepAliveOk, appUpsertOk, appKeepAliveOk),
		deny(appUpsertErr, appKeepAliveErr, handlerClose),
	)

	// set up to induce some failures, but not enough to cause the control
	// stream to be closed.
	auth.mu.Lock()
	auth.failUpserts = 1
	auth.failKeepAlives = 2
	auth.mu.Unlock()

	// keepalive should fail twice, but since the upsert is already known
	// to have succeeded, we should not see an upsert failure yet.
	awaitEvents(t, events,
		expect(appKeepAliveErr, appKeepAliveErr),
		deny(appUpsertErr, handlerClose),
	)

	// jitter can sometimes cause one app to keepalive twice before another has completed one keepalive. for that
	// reason, we want 2x the number of apps worth of keepalives to ensure that the failed keepalive counts associated
	// with each app have been reset. otherwise, later parts of this test become flaky.
	var keepaliveEvents []testEvent
	for range appCount {
		keepaliveEvents = append(keepaliveEvents, []testEvent{appKeepAliveOk, appKeepAliveOk}...)
	}

	awaitEvents(t, events,
		expect(keepaliveEvents...),
		deny(appKeepAliveErr, handlerClose),
	)

	for i := range appCount {
		err := downstream.Send(ctx, &proto.InventoryHeartbeat{
			AppServer: &types.AppServerV3{
				Metadata: types.Metadata{
					Name: serverID,
				},
				Spec: types.AppServerSpecV3{
					HostID: serverID,
					App: &types.AppV3{
						Kind:    types.KindApp,
						Version: types.V3,
						Metadata: types.Metadata{
							Name: fmt.Sprintf("app-%d", i),
						},
						Spec: types.AppSpecV3{},
					},
				},
			},
		})
		require.NoError(t, err)
	}

	// we should now see an upsert failure, but no additional
	// keepalive failures, and the upsert should succeed on retry.
	awaitEvents(t, events,
		expect(appUpsertErr, appUpsertRetryOk),
		deny(appKeepAliveErr, handlerClose),
	)

	// limit time of ping call
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// execute ping
	_, err := handle.Ping(pingCtx, 1)
	require.NoError(t, err)

	// ensure that local app keepalive states have reset to healthy by waiting
	// on a full cycle+ worth of keepalives without errors.
	awaitEvents(t, events,
		expect(keepAliveAppTick, keepAliveAppTick),
		deny(appKeepAliveErr, handlerClose),
	)

	// set up to induce enough consecutive keepalive errors to cause removal
	// of server-side keepalive state.
	auth.mu.Lock()
	auth.failKeepAlives = 3 * appCount
	auth.mu.Unlock()

	// expect that all app keepalives fail, then the app is removed.
	var expectedEvents []testEvent
	for range appCount {
		expectedEvents = append(expectedEvents, []testEvent{appKeepAliveErr, appKeepAliveErr, appKeepAliveErr, appKeepAliveDel}...)
	}

	// wait for failed keepalives to trigger removal
	awaitEvents(t, events,
		expect(expectedEvents...),
		deny(handlerClose),
	)

	// set up to induce enough consecutive errors to cause stream closure
	auth.mu.Lock()
	auth.failUpserts = 5
	auth.mu.Unlock()

	err = downstream.Send(ctx, &proto.InventoryHeartbeat{
		AppServer: &types.AppServerV3{
			Metadata: types.Metadata{
				Name: serverID,
			},
			Spec: types.AppServerSpecV3{
				HostID: serverID,
				App: &types.AppV3{
					Kind:     types.KindApp,
					Version:  types.V3,
					Metadata: types.Metadata{},
					Spec:     types.AppSpecV3{},
				},
			},
		},
	})
	require.NoError(t, err)

	// both the initial upsert and the retry should fail, then the handle should
	// close.
	awaitEvents(t, events,
		expect(appUpsertErr, appUpsertRetryErr, handlerClose),
		deny(appUpsertOk),
	)

	// verify that closure propagates to server and client side interfaces
	closeTimeout := time.After(time.Second * 10)
	select {
	case <-handle.Done():
	case <-closeTimeout:
		t.Fatal("timeout waiting for handle closure")
	}
	select {
	case <-downstream.Done():
	case <-closeTimeout:
		t.Fatal("timeout waiting for handle closure")
	}

	// verify that hb counter has been decremented (counter is decremented concurrently, but
	// always *before* closure is propagated to downstream handle, hence being safe to load
	// here).
	require.Equal(t, int64(0), controller.instanceHBVariableDuration.Count())

	// verify that metrics have been updated correctly
	require.Zero(t, rc.count())
}

// TestDatabaseServerBasics verifies basic expected behaviors for a single control stream heartbeating
// a database server.
func TestDatabaseServerBasics(t *testing.T) {
	const serverID = "test-server"
	const dbCount = 3

	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan testEvent, 1024)

	auth := &fakeAuth{}

	rc := &resourceCounter{}
	controller := NewController(
		auth,
		usagereporter.DiscardUsageReporter{},
		withServerKeepAlive(time.Millisecond*200),
		withTestEventsChannel(events),
		WithOnConnect(rc.onConnect),
		WithOnDisconnect(rc.onDisconnect),
	)
	defer controller.Close()

	// set up fake in-memory control stream
	upstream, downstream := client.InventoryControlStreamPipe()
	t.Cleanup(func() {
		controller.Close()
		upstream.Close()
		downstream.Close()
	})

	// launch goroutine to respond to ping requests
	go func() {
		for {
			select {
			case msg := <-downstream.Recv():
				downstream.Send(ctx, &proto.UpstreamInventoryPong{
					ID: msg.(*proto.DownstreamInventoryPing).GetID(),
				})
			case <-downstream.Done():
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	controller.RegisterControlStream(upstream, &proto.UpstreamInventoryHello{
		ServerID: serverID,
		Version:  teleport.Version,
		Services: types.SystemRoles{types.RoleDatabase}.StringSlice(),
	})

	// verify that control stream handle is now accessible
	handle, ok := controller.GetControlStream(serverID)
	require.True(t, ok)

	// verify that hb counter has been incremented
	require.Equal(t, int64(1), controller.instanceHBVariableDuration.Count())

	// send a fake db server heartbeat
	for i := range dbCount {
		err := downstream.Send(ctx, &proto.InventoryHeartbeat{
			DatabaseServer: &types.DatabaseServerV3{
				Metadata: types.Metadata{
					Name: serverID,
				},
				Spec: types.DatabaseServerSpecV3{
					HostID:   serverID,
					Hostname: serverID,
					Database: &types.DatabaseV3{
						Kind:    types.KindDatabase,
						Version: types.V3,
						Metadata: types.Metadata{
							Name: fmt.Sprintf("db-%d", i),
						},
						Spec: types.DatabaseSpecV3{},
					},
				},
			},
		})
		require.NoError(t, err)
	}

	// verify that heartbeat creates both an upsert and a keepalive
	awaitEvents(t, events,
		expect(dbUpsertOk, dbKeepAliveOk, dbUpsertOk, dbKeepAliveOk, dbUpsertOk, dbKeepAliveOk),
		deny(dbUpsertErr, dbKeepAliveErr, handlerClose),
	)

	// set up to induce delete failure
	auth.mu.Lock()
	auth.failDeletes = 1
	auth.mu.Unlock()

	// stop a heartbeat
	err := downstream.Send(ctx, &proto.UpstreamInventoryStopHeartbeat{
		Kind: proto.StopHeartbeatKind_STOP_HEARTBEAT_KIND_DATABASE_SERVER,
		Name: "db-1",
	})
	require.NoError(t, err)

	// verify that keep alive stops, even if the heartbeat couldn't be deleted
	awaitEvents(t, events,
		expect(dbKeepAliveDel, dbDelErr),
		deny(dbDelOk, dbUpsertErr, dbKeepAliveErr, handlerClose),
	)
	require.Equal(t, dbCount-1, rc.count())

	// verify heartbeat stop keepalive idempotency
	err = downstream.Send(ctx, &proto.UpstreamInventoryStopHeartbeat{
		Kind: proto.StopHeartbeatKind_STOP_HEARTBEAT_KIND_DATABASE_SERVER,
		Name: "db-1",
	})
	require.NoError(t, err)
	require.Equal(t, dbCount-1, rc.count())

	awaitEvents(t, events,
		expect(dbStopErr),
		deny(dbKeepAliveDel, dbDelOk, dbDelErr, dbUpsertErr, dbKeepAliveErr, handlerClose),
	)
	require.Equal(t, dbCount-1, rc.count())

	// set up to induce some failures, but not enough to cause the control
	// stream to be closed.
	auth.mu.Lock()
	auth.failUpserts = 1
	auth.failKeepAlives = 2
	auth.mu.Unlock()

	// keepalive should fail twice, but since the upsert is already known
	// to have succeeded, we should not see an upsert failure yet.
	awaitEvents(t, events,
		expect(dbKeepAliveErr, dbKeepAliveErr),
		deny(dbUpsertErr, handlerClose),
	)

	// jitter can sometimes cause one app to keepalive twice before another has completed one keepalive. for that
	// reason, we want 2x the number of apps worth of keepalives to ensure that the failed keepalive counts associated
	// with each app have been reset. otherwise, later parts of this test become flaky.
	var keepaliveEvents []testEvent
	for range dbCount {
		keepaliveEvents = append(keepaliveEvents, []testEvent{dbKeepAliveOk, dbKeepAliveOk}...)
	}

	awaitEvents(t, events,
		expect(keepaliveEvents...),
		deny(appKeepAliveErr, handlerClose),
	)

	for i := range dbCount {
		err := downstream.Send(ctx, &proto.InventoryHeartbeat{
			DatabaseServer: &types.DatabaseServerV3{
				Metadata: types.Metadata{
					Name: serverID,
				},
				Spec: types.DatabaseServerSpecV3{
					HostID: serverID,
					Database: &types.DatabaseV3{
						Kind:    types.KindDatabase,
						Version: types.V3,
						Metadata: types.Metadata{
							Name: fmt.Sprintf("db-%d", i),
						},
						Spec: types.DatabaseSpecV3{},
					},
				},
			},
		})
		require.NoError(t, err)
	}

	// we should now see an upsert failure, but no additional
	// keepalive failures, and the upsert should succeed on retry.
	awaitEvents(t, events,
		expect(dbUpsertErr, dbUpsertRetryOk),
		deny(dbKeepAliveErr, handlerClose),
	)

	// limit time of ping call
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// execute ping
	_, err = handle.Ping(pingCtx, 1)
	require.NoError(t, err)

	// ensure that local db keepalive states have reset to healthy by waiting
	// on a full cycle+ worth of keepalives without errors.
	awaitEvents(t, events,
		expect(keepAliveDatabaseTick, keepAliveDatabaseTick),
		deny(dbKeepAliveErr, handlerClose),
	)

	// set up to induce enough consecutive keepalive errors to cause removal
	// of server-side keepalive state.
	auth.mu.Lock()
	auth.failKeepAlives = 3 * dbCount
	auth.mu.Unlock()

	// expect that all db keepalives fail, then the db is removed.
	var expectedEvents []testEvent
	for range dbCount {
		expectedEvents = append(expectedEvents, []testEvent{dbKeepAliveErr, dbKeepAliveErr, dbKeepAliveErr, dbKeepAliveDel}...)
	}

	// wait for failed keepalives to trigger removal
	awaitEvents(t, events,
		expect(expectedEvents...),
		deny(handlerClose),
	)

	// set up to induce enough consecutive errors to cause stream closure
	auth.mu.Lock()
	auth.failUpserts = 5
	auth.mu.Unlock()

	err = downstream.Send(ctx, &proto.InventoryHeartbeat{
		DatabaseServer: &types.DatabaseServerV3{
			Metadata: types.Metadata{
				Name: serverID,
			},
			Spec: types.DatabaseServerSpecV3{
				HostID: serverID,
				Database: &types.DatabaseV3{
					Kind:     types.KindDatabase,
					Version:  types.V3,
					Metadata: types.Metadata{},
					Spec:     types.DatabaseSpecV3{},
				},
			},
		},
	})
	require.NoError(t, err)

	// both the initial upsert and the retry should fail, then the handle should
	// close.
	awaitEvents(t, events,
		expect(dbUpsertErr, dbUpsertRetryErr, handlerClose),
		deny(dbUpsertOk),
	)

	// verify that closure propagates to server and client side interfaces
	closeTimeout := time.After(time.Second * 10)
	select {
	case <-handle.Done():
	case <-closeTimeout:
		t.Fatal("timeout waiting for handle closure")
	}
	select {
	case <-downstream.Done():
	case <-closeTimeout:
		t.Fatal("timeout waiting for handle closure")
	}

	// verify that hb counter has been decremented (counter is decremented concurrently, but
	// always *before* closure is propagated to downstream handle, hence being safe to load
	// here).
	require.Equal(t, int64(0), controller.instanceHBVariableDuration.Count())

	// verify that metrics have been updated correctly
	require.Zero(t, rc.count())
}

// TestInstanceHeartbeat verifies basic expected behaviors for instance heartbeat.
func TestInstanceHeartbeat_Disabled(t *testing.T) {
	const serverID = "test-instance"
	const peerAddr = "1.2.3.4:456"

	events := make(chan testEvent, 1024)

	auth := &fakeAuth{}

	controller := NewController(
		auth,
		usagereporter.DiscardUsageReporter{},
		withInstanceHBInterval(time.Millisecond*200),
		withTestEventsChannel(events),
	)
	defer controller.Close()

	// set up fake in-memory control stream
	upstream, _ := client.InventoryControlStreamPipe(client.ICSPipePeerAddr(peerAddr))

	controller.RegisterControlStream(upstream, &proto.UpstreamInventoryHello{
		ServerID: serverID,
		Version:  teleport.Version,
		Services: types.SystemRoles{types.RoleNode}.StringSlice(),
	})

	// verify that control stream handle is now accessible
	_, ok := controller.GetControlStream(serverID)
	require.True(t, ok)

	// verify that no instance heartbeats are emitted
	awaitEvents(t, events,
		deny(instanceHeartbeatOk, instanceHeartbeatErr, instanceCompareFailed, handlerClose),
	)
}

func TestInstanceHeartbeatDisabledEnv(t *testing.T) {
	t.Setenv("TELEPORT_UNSTABLE_DISABLE_INSTANCE_HB", "yes")

	controller := NewController(
		&fakeAuth{},
		usagereporter.DiscardUsageReporter{},
	)
	defer controller.Close()

	require.False(t, controller.instanceHBEnabled)
}

// TestInstanceHeartbeat verifies basic expected behaviors for instance heartbeat.
func TestInstanceHeartbeat(t *testing.T) {
	const serverID = "test-instance"
	const peerAddr = "1.2.3.4:456"

	events := make(chan testEvent, 1024)

	auth := &fakeAuth{}

	controller := NewController(
		auth,
		usagereporter.DiscardUsageReporter{},
		withInstanceHBInterval(time.Millisecond*200),
		withTestEventsChannel(events),
	)

	// set up fake in-memory control stream
	upstream, downstream := client.InventoryControlStreamPipe(client.ICSPipePeerAddr(peerAddr))
	t.Cleanup(func() {
		controller.Close()
		downstream.Close()
		upstream.Close()
	})

	// Launch goroutine to consume downstream request and don't block control steam handler.
	go func() {
		for {
			select {
			case <-downstream.Recv():
			case <-downstream.Done():
				return
			}
		}
	}()

	controller.RegisterControlStream(upstream, &proto.UpstreamInventoryHello{
		ServerID: serverID,
		Version:  teleport.Version,
		Services: types.SystemRoles{types.RoleNode, types.RoleApp}.StringSlice(),
	})

	// verify that control stream handle is now accessible
	handle, ok := controller.GetControlStream(serverID)
	require.True(t, ok)

	// verify that instance heartbeat succeeds
	awaitEvents(t, events,
		expect(instanceHeartbeatOk),
		deny(instanceHeartbeatErr, instanceCompareFailed, handlerClose),
	)

	// verify the service counter shows the correct number for the given services.
	require.Equal(t, uint64(1), controller.serviceCounter.get(types.RoleNode))
	require.Equal(t, uint64(1), controller.serviceCounter.get(types.RoleApp))

	// this service was not seen, so it should be 0.
	require.Equal(t, uint64(0), controller.serviceCounter.get(types.RoleOkta))

	// set up single failure of upsert. stream should recover.
	auth.mu.Lock()
	auth.failUpsertInstance = 1
	auth.mu.Unlock()

	// verify that heartbeat error occurs
	awaitEvents(t, events,
		expect(instanceHeartbeatErr),
		deny(instanceCompareFailed, handlerClose),
	)

	// verify that recovery happens
	awaitEvents(t, events,
		expect(instanceHeartbeatOk),
		deny(instanceHeartbeatErr, instanceCompareFailed, handlerClose),
	)

	// verify the service counter shows the correct number for the given services.
	require.Equal(t, map[types.SystemRole]uint64{
		types.RoleApp:  1,
		types.RoleNode: 1,
	}, controller.ConnectedServiceCounts())
	require.Equal(t, uint64(1), controller.ConnectedServiceCount(types.RoleNode))
	require.Equal(t, uint64(1), controller.ConnectedServiceCount(types.RoleApp))

	// set up double failure of CAS. stream should not recover.
	auth.mu.Lock()
	auth.failUpsertInstance = 2
	auth.mu.Unlock()

	// expect failure and handle closure
	awaitEvents(t, events,
		expect(instanceHeartbeatErr, handlerClose),
	)

	// verify that closure propagates to server and client side interfaces
	closeTimeout := time.After(time.Second * 10)
	select {
	case <-handle.Done():
	case <-closeTimeout:
		t.Fatal("timeout waiting for handle closure")
	}
	select {
	case <-downstream.Done():
	case <-closeTimeout:
		t.Fatal("timeout waiting for handle closure")
	}

	// verify the service counter now shows no connected services.
	require.Equal(t, map[types.SystemRole]uint64{
		types.RoleApp:  0,
		types.RoleNode: 0,
	}, controller.ConnectedServiceCounts())
	require.Equal(t, uint64(0), controller.ConnectedServiceCount(types.RoleNode))
	require.Equal(t, uint64(0), controller.ConnectedServiceCount(types.RoleApp))
}

// TestUpdateLabels verifies that an instance's labels can be updated over an
// inventory control stream.
func TestUpdateLabels(t *testing.T) {
	t.Parallel()
	const serverID = "test-instance"
	const peerAddr = "1.2.3.4:456"

	events := make(chan testEvent, 1024)

	auth := &fakeAuth{}

	controller := NewController(
		auth,
		usagereporter.DiscardUsageReporter{},
		withInstanceHBInterval(time.Millisecond*200),
		withTestEventsChannel(events),
	)
	defer controller.Close()

	// Set up fake in-memory control stream.
	upstream, downstream := client.InventoryControlStreamPipe(client.ICSPipePeerAddr(peerAddr))
	upstreamHello := &proto.UpstreamInventoryHello{
		ServerID: serverID,
		Version:  teleport.Version,
		Services: types.SystemRoles{types.RoleNode}.StringSlice(),
	}
	downstreamHello := &proto.DownstreamInventoryHello{
		Version:  teleport.Version,
		ServerID: "auth",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	downstreamHandle, err := NewDownstreamHandle(func(ctx context.Context) (client.DownstreamInventoryControlStream, error) {
		return downstream, nil
	}, func(_ context.Context) (*proto.UpstreamInventoryHello, error) { return upstreamHello, nil })
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, downstreamHandle.Close()) })

	// Wait for upstream hello.
	select {
	case msg := <-upstream.Recv():
		require.Equal(t, upstreamHello, msg)
	case <-ctx.Done():
		require.Fail(t, "never got upstream hello")
	}
	require.NoError(t, upstream.Send(ctx, downstreamHello))
	controller.RegisterControlStream(upstream, upstreamHello)

	// Verify that control stream upstreamHandle is now accessible.
	upstreamHandle, ok := controller.GetControlStream(serverID)
	require.True(t, ok)

	// Update labels.
	labels := map[string]string{"a": "1", "b": "2"}
	require.NoError(t, upstreamHandle.UpdateLabels(ctx, proto.LabelUpdateKind_SSHServerCloudLabels, labels))

	require.Eventually(t, func() bool {
		require.Equal(t, labels, downstreamHandle.GetUpstreamLabels(proto.LabelUpdateKind_SSHServerCloudLabels))
		return true
	}, time.Second, 100*time.Millisecond)
}

// TestAgentMetadata verifies that an instance's agent metadata is received in
// inventory control stream.
func TestAgentMetadata(t *testing.T) {
	t.Parallel()

	const serverID = "test-instance"
	const peerAddr = "1.2.3.4:456"

	events := make(chan testEvent, 1024)

	auth := &fakeAuth{}

	controller := NewController(
		auth,
		usagereporter.DiscardUsageReporter{},
		withInstanceHBInterval(time.Millisecond*200),
		withTestEventsChannel(events),
	)
	defer controller.Close()

	// Set up fake in-memory control stream.
	upstream, downstream := client.InventoryControlStreamPipe(client.ICSPipePeerAddr(peerAddr))
	upstreamHello := &proto.UpstreamInventoryHello{
		ServerID: serverID,
		Version:  teleport.Version,
		Services: types.SystemRoles{types.RoleNode}.StringSlice(),
	}
	downstreamHello := &proto.DownstreamInventoryHello{
		Version:  teleport.Version,
		ServerID: "auth",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	inventoryHandle, err := NewDownstreamHandle(
		func(ctx context.Context) (client.DownstreamInventoryControlStream, error) {
			return downstream, nil
		},
		func(ctx context.Context) (*proto.UpstreamInventoryHello, error) { return upstreamHello, nil },
		withMetadataGetter(func(ctx context.Context) (*metadata.Metadata, error) {
			return &metadata.Metadata{
				OS:                    "llamaOS",
				OSVersion:             "1.2.3",
				HostArchitecture:      "llama",
				GlibcVersion:          "llama.5.6.7",
				InstallMethods:        []string{"llama", "alpaca"},
				ContainerRuntime:      "test",
				ContainerOrchestrator: "test",
				CloudEnvironment:      "llama-cloud",
			}, nil
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, inventoryHandle.Close()) })

	// Wait for upstream hello.
	select {
	case msg := <-upstream.Recv():
		require.Equal(t, upstreamHello, msg)
	case <-ctx.Done():
		require.Fail(t, "never got upstream hello")
	}
	require.NoError(t, upstream.Send(ctx, downstreamHello))
	controller.RegisterControlStream(upstream, upstreamHello)

	// Verify that control stream upstreamHandle is now accessible.
	upstreamHandle, ok := controller.GetControlStream(serverID)
	require.True(t, ok)

	// Validate that the agent's metadata ends up in the auth server.
	require.Eventually(t, func() bool {
		return slices.Equal([]string{"llama", "alpaca"}, upstreamHandle.AgentMetadata().InstallMethods) &&
			upstreamHandle.AgentMetadata().OS == "llamaOS"
	}, 10*time.Second, 200*time.Millisecond)
}

func TestGoodbye(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		supportsGoodbye bool
		deleteResources bool
		softReload      bool
	}{
		{
			name: "no goodbye",
		},
		{
			name:            "goodbye termination",
			deleteResources: true,
			supportsGoodbye: true,
		},
		{
			name:            "goodbye soft-reload",
			supportsGoodbye: true,
			softReload:      true,
		},
	}

	upstreamHello := &proto.UpstreamInventoryHello{
		ServerID: "llama",
		Version:  teleport.Version,
		Services: types.SystemRoles{types.RoleNode, types.RoleApp}.StringSlice(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Test Setup: crafting a controller and an upstream/downstream stream pipe
			controller := NewController(
				&fakeAuth{},
				usagereporter.DiscardUsageReporter{},
				withInstanceHBInterval(time.Millisecond*200),
			)
			defer controller.Close()
			upstream, downstream := client.InventoryControlStreamPipe(client.ICSPipePeerAddr("127.0.0.1:8090"))

			currentDownstream := downstream
			downstreamHello := &proto.DownstreamInventoryHello{
				Version:  teleport.Version,
				ServerID: "auth",
				Capabilities: &proto.DownstreamInventoryHello_SupportedCapabilities{
					AppCleanup:     test.supportsGoodbye,
					AppHeartbeats:  true,
					NodeHeartbeats: true,
				},
			}

			// Test setup: creating downstream handler
			clock := clockwork.NewFakeClock()
			handle, err := NewDownstreamHandle(
				func(ctx context.Context) (client.DownstreamInventoryControlStream, error) {
					return currentDownstream, nil
				}, func(ctx context.Context) (*proto.UpstreamInventoryHello, error) {
					return upstreamHello, nil
				}, WithDownstreamClock(clock))
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, handle.Close())
			})

			// Test setup: wait for the downstream handler to finish its startup and respond to its hello
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			select {
			case msg := <-upstream.Recv():
				require.Equal(t, upstreamHello, msg)
			case <-ctx.Done():
				require.Fail(t, "never got upstream hello")
			}
			require.NoError(t, upstream.Send(ctx, downstreamHello))

			// Test execution part 1: Validating that calling SetAndSendGoodbye does
			// cause the downstream handler t send a goodbye.

			// Fire a routine that will call SetAndSendGoodbye
			// Then send a Pong message to indicate the test is done.
			nonce := rand.Int()
			go func() {
				ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				handle.SetAndSendGoodbye(ctx, test.deleteResources, test.softReload)

				// After calling the goodbye, we do a little hack by sending a pong
				// This will allow us to know that this routine is done on the other side of the test
				// while making sure we are strictly after the goodbye message
				// (SendGoodBye is synchronous, and we send messages over the same channel)
				select {
				case sender := <-handle.Sender():
					sender.Send(ctx, &proto.UpstreamInventoryPong{ID: uint64(nonce)})
				case <-ctx.Done():
					assert.Fail(t, "never got downstream sender, was not able to send 'end-of-test pong'")
				}
			}()

			// Wait until we receive the pong.
			var receivedGoodbye *proto.UpstreamInventoryGoodbye
			timeoutC := time.After(10 * time.Second)
		OuterLoop:
			for {
				select {
				case msg := <-upstream.Recv():
					switch msg := msg.(type) {
					case *proto.UpstreamInventoryHello, *proto.InventoryHeartbeat,
						*proto.UpstreamInventoryAgentMetadata:
						// We skip all usual messages
					case *proto.UpstreamInventoryGoodbye:
						// We register the goodbye
						receivedGoodbye = msg
					case *proto.UpstreamInventoryPong:
						// The emitter routine is done with its part, we stop waiting for events
						require.Equal(t, nonce, int(msg.ID), "received pong message without the 'end-of-test ID'")
						break OuterLoop
					default:
						require.Fail(t, "unexpected message type", msg)
					}
				case <-upstream.Done():
					require.Fail(t, "upstream unexpectedly closed")
				case <-timeoutC:
					require.Fail(t, "timed out waiting for 'end-of-test pong'")
				}
			}

			// Test validation pt.1: Check if we received a pong
			if test.supportsGoodbye {
				require.NotNil(t, receivedGoodbye)
				require.Equal(t, test.deleteResources, receivedGoodbye.DeleteResources)
				require.Equal(t, test.softReload, receivedGoodbye.SoftReload)
			} else {
				require.Nil(t, receivedGoodbye)
			}

			// Don't test further if we don't support goodbye
			if !test.supportsGoodbye {
				return
			}

			// Test execution pt.2: See that the downstreamHandler sends back the goodbye
			// when the stream is re-established.

			newUpstream, newDownstream := client.InventoryControlStreamPipe(client.ICSPipePeerAddr("127.0.0.1:8090"))
			currentDownstream = newDownstream

			// Simulating a failure on the old control stream, asking for a reconnection
			require.NoError(t, downstream.CloseWithError(io.EOF))

			ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			// Wait to hit the handle linear retry, then jump into the future
			require.NoError(t, clock.BlockUntilContext(ctx, 1))
			select {
			case <-ctx.Done():
				require.Fail(t, "upstream never waited")
			default:
				clock.Advance(time.Minute)
			}

			// Test setup: wait for the downstream handler to finish its startup and respond to its hello
			select {
			case msg := <-newUpstream.Recv():
				require.Equal(t, upstreamHello, msg)
			case <-ctx.Done():
				require.Fail(t, "never got upstream hello")
			}
			require.NoError(t, newUpstream.Send(ctx, downstreamHello))

			// Test execution: wait until we receive at least 1 goodbye.
			timeoutC = time.After(10 * time.Second)
			for {
				select {
				case msg := <-newUpstream.Recv():
					switch msg := msg.(type) {
					case *proto.UpstreamInventoryHello, *proto.InventoryHeartbeat,
						*proto.UpstreamInventoryPong, *proto.UpstreamInventoryAgentMetadata:
					case *proto.UpstreamInventoryGoodbye:
						require.Empty(t, cmp.Diff(receivedGoodbye, msg, protocmp.Transform()), "re-emitted goodbye should be identical")
						return
					default:
						require.Fail(t, "unexpected message type", msg)
					}

				case <-newUpstream.Done():
					require.Fail(t, "upstream unexpectedly closed")
				case <-timeoutC:
					require.Fail(t, "timeout waiting for goodbye message")
				}
			}

		})
	}
}

func TestKubernetesServerBasics(t *testing.T) {
	const serverID = "test-server"
	const kubeCount = 3

	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan testEvent, 1024)

	auth := &fakeAuth{}

	rc := &resourceCounter{}
	controller := NewController(
		auth,
		usagereporter.DiscardUsageReporter{},
		withServerKeepAlive(time.Millisecond*200),
		withTestEventsChannel(events),
		WithOnConnect(rc.onConnect),
		WithOnDisconnect(rc.onDisconnect),
	)
	defer controller.Close()

	// set up fake in-memory control stream
	upstream, downstream := client.InventoryControlStreamPipe()
	// launch goroutine to respond to ping requests
	go func() {
		for {
			select {
			case msg := <-downstream.Recv():
				downstream.Send(ctx, &proto.UpstreamInventoryPong{
					ID: msg.(*proto.DownstreamInventoryPing).GetID(),
				})
			case <-downstream.Done():
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	controller.RegisterControlStream(upstream, &proto.UpstreamInventoryHello{
		ServerID: serverID,
		Version:  teleport.Version,
		Services: types.SystemRoles{types.RoleKube}.StringSlice(),
	})

	// verify that control stream handle is now accessible
	handle, ok := controller.GetControlStream(serverID)
	require.True(t, ok)

	// verify that hb counter has been incremented
	require.Equal(t, int64(1), controller.instanceHBVariableDuration.Count())

	// send a fake kube server heartbeat
	for i := range kubeCount {
		err := downstream.Send(ctx, &proto.InventoryHeartbeat{
			KubernetesServer: &types.KubernetesServerV3{
				Metadata: types.Metadata{
					Name: serverID,
				},
				Spec: types.KubernetesServerSpecV3{
					HostID:   serverID,
					Hostname: serverID,
					Cluster: &types.KubernetesClusterV3{
						Kind:    types.KindKubernetesCluster,
						Version: types.V3,
						Metadata: types.Metadata{
							Name: fmt.Sprintf("cluster-%d", i),
						},
						Spec: types.KubernetesClusterSpecV3{},
					},
				},
			},
		})
		require.NoError(t, err)
	}

	// verify that heartbeat creates both an upsert and a keepalive
	awaitEvents(t, events,
		expect(kubeUpsertOk, kubeKeepAliveOk, kubeUpsertOk, kubeKeepAliveOk, kubeUpsertOk, kubeKeepAliveOk),
		deny(kubeUpsertErr, kubeKeepAliveErr, handlerClose),
	)

	// set up to induce some failures, but not enough to cause the control
	// stream to be closed.
	auth.mu.Lock()
	auth.failUpserts = 1
	auth.failKeepAlives = 2
	auth.mu.Unlock()

	// keepalive should fail twice, but since the upsert is already known
	// to have succeeded, we should not see an upsert failure yet.
	awaitEvents(t, events,
		expect(kubeKeepAliveErr, kubeKeepAliveErr),
		deny(kubeUpsertErr, handlerClose),
	)

	// jitter can sometimes cause one app to keepalive twice before another has completed one keepalive. for that
	// reason, we want 2x the number of apps worth of keepalives to ensure that the failed keepalive counts associated
	// with each app have been reset. otherwise, later parts of this test become flaky.
	var keepaliveEvents []testEvent
	for range kubeCount {
		keepaliveEvents = append(keepaliveEvents, []testEvent{kubeKeepAliveOk, kubeKeepAliveOk}...)
	}

	awaitEvents(t, events,
		expect(keepaliveEvents...),
		deny(appKeepAliveErr, handlerClose),
	)

	for i := range kubeCount {
		err := downstream.Send(ctx, &proto.InventoryHeartbeat{
			KubernetesServer: &types.KubernetesServerV3{
				Metadata: types.Metadata{
					Name: serverID,
				},
				Spec: types.KubernetesServerSpecV3{
					HostID:   serverID,
					Hostname: serverID,
					Cluster: &types.KubernetesClusterV3{
						Kind:    types.KindKubernetesCluster,
						Version: types.V3,
						Metadata: types.Metadata{
							Name: fmt.Sprintf("cluster-%d", i),
						},
						Spec: types.KubernetesClusterSpecV3{},
					},
				},
			},
		})
		require.NoError(t, err)
	}

	// we should now see an upsert failure, but no additional
	// keepalive failures, and the upsert should succeed on retry.
	awaitEvents(t, events,
		expect(kubeUpsertErr, kubeUpsertRetryOk),
		deny(kubeKeepAliveErr, handlerClose),
	)

	// limit time of ping call
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// execute ping
	_, err := handle.Ping(pingCtx, 1)
	require.NoError(t, err)

	// ensure that local app keepalive states have reset to healthy by waiting
	// on a full cycle+ worth of keepalives without errors.
	awaitEvents(t, events,
		expect(keepAliveKubeTick, keepAliveKubeTick),
		deny(kubeKeepAliveErr, handlerClose),
	)

	// set up to induce enough consecutive keepalive errors to cause removal
	// of server-side keepalive state.
	auth.mu.Lock()
	auth.failKeepAlives = 3 * kubeCount
	auth.mu.Unlock()

	// expect that all app keepalives fail, then the app is removed.
	var expectedEvents []testEvent
	for range kubeCount {
		expectedEvents = append(expectedEvents, []testEvent{kubeKeepAliveErr, kubeKeepAliveErr, kubeKeepAliveErr, kubeKeepAliveDel}...)
	}

	// wait for failed keepalives to trigger removal
	awaitEvents(t, events,
		expect(expectedEvents...),
		deny(handlerClose),
	)

	// set up to induce enough consecutive errors to cause stream closure
	auth.mu.Lock()
	auth.failUpserts = 5
	auth.mu.Unlock()

	err = downstream.Send(ctx, &proto.InventoryHeartbeat{
		KubernetesServer: &types.KubernetesServerV3{
			Metadata: types.Metadata{
				Name: serverID,
			},
			Spec: types.KubernetesServerSpecV3{
				HostID:   serverID,
				Hostname: serverID,
				Cluster: &types.KubernetesClusterV3{
					Kind:    types.KindKubernetesCluster,
					Version: types.V3,
					Metadata: types.Metadata{
						Name: "cluster-1",
					},
					Spec: types.KubernetesClusterSpecV3{},
				},
			},
		},
	})
	require.NoError(t, err)

	// both the initial upsert and the retry should fail, then the handle should
	// close.
	awaitEvents(t, events,
		expect(kubeUpsertErr, kubeUpsertRetryErr, handlerClose),
		deny(kubeUpsertOk),
	)

	// verify that closure propagates to server and client side interfaces
	closeTimeout := time.After(time.Second * 10)
	select {
	case <-handle.Done():
	case <-closeTimeout:
		t.Fatal("timeout waiting for handle closure")
	}
	select {
	case <-downstream.Done():
	case <-closeTimeout:
		t.Fatal("timeout waiting for handle closure")
	}

	// verify that hb counter has been decremented (counter is decremented concurrently, but
	// always *before* closure is propagated to downstream handle, hence being safe to load
	// here).
	require.Equal(t, int64(0), controller.instanceHBVariableDuration.Count())

	// verify that metrics have been updated correctly
	require.Zero(t, rc.count())
}

func TestGetSender(t *testing.T) {
	controller := NewController(
		&fakeAuth{},
		usagereporter.DiscardUsageReporter{},
		withInstanceHBInterval(time.Millisecond*200),
	)
	defer controller.Close()

	// Set up fake in-memory control stream.
	upstream, downstream := client.InventoryControlStreamPipe(client.ICSPipePeerAddr("127.0.0.1:8090"))

	downstreamHello := &proto.DownstreamInventoryHello{
		Version:  teleport.Version,
		ServerID: "auth",
		Capabilities: &proto.DownstreamInventoryHello_SupportedCapabilities{
			AppCleanup:     true,
			AppHeartbeats:  true,
			NodeHeartbeats: true,
		},
	}

	upstreamHello := &proto.UpstreamInventoryHello{
		ServerID: "llama",
		Version:  teleport.Version,
		Services: types.SystemRoles{types.RoleNode, types.RoleApp}.StringSlice(),
	}

	handle, err := NewDownstreamHandle(func(ctx context.Context) (client.DownstreamInventoryControlStream, error) {
		return downstream, nil
	}, func(ctx context.Context) (*proto.UpstreamInventoryHello, error) { return upstreamHello, nil })
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, handle.Close()) })

	// Validate that the sender is not present prior to
	// the stream becoming healthy.
	s, ok := handle.GetSender()
	require.False(t, ok)
	require.Nil(t, s)

	// Wait for upstream hello.
	select {
	case msg := <-upstream.Recv():
		require.Equal(t, upstreamHello, msg)
	case <-time.After(5 * time.Second):
		require.Fail(t, "never got upstream hello")
	}
	// Send the downstream hello so that the
	// sender becomes available.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, upstream.Send(ctx, downstreamHello))

	// Validate that once healthy the sender is provided.
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		s, ok = handle.GetSender()
		assert.True(t, ok)
		assert.NotNil(t, s)
	}, 10*time.Second, 100*time.Millisecond)
}

// TestTimeReconciliation verifies basic behavior of the time reconciliation check.
func TestTimeReconciliation(t *testing.T) {
	const serverID = "test-server"
	const peerAddr = "1.2.3.4:456"
	const wantAddr = "1.2.3.4:123"

	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan testEvent, 1024)
	auth := &fakeAuth{
		expectAddr: wantAddr,
	}

	clock := clockwork.NewRealClock()
	controller := NewController(
		auth,
		usagereporter.DiscardUsageReporter{},
		withInstanceHBInterval(time.Millisecond*200),
		withTestEventsChannel(events),
		WithClock(clock),
	)

	// Set up fake in-memory control stream.
	upstream, downstream := client.InventoryControlStreamPipe(client.ICSPipePeerAddr(peerAddr))

	t.Cleanup(func() {
		require.NoError(t, downstream.Close())
		require.NoError(t, upstream.Close())
		require.NoError(t, controller.Close())
		cancel()
	})

	// Launch goroutine to respond to clock request.
	go func() {
		for {
			select {
			case msg := <-downstream.Recv():
				downstream.Send(ctx, &proto.UpstreamInventoryPong{
					ID:          msg.(*proto.DownstreamInventoryPing).GetID(),
					SystemClock: timestamppb.New(clock.Now().Add(-time.Minute)),
				})
			case <-downstream.Done():
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	controller.RegisterControlStream(upstream, &proto.UpstreamInventoryHello{
		ServerID: serverID,
		Version:  teleport.Version,
		Services: types.SystemRoles{types.RoleNode}.StringSlice(),
	})

	_, ok := controller.GetControlStream(serverID)
	require.True(t, ok)

	awaitEvents(t, events, expect(pongOk))
	awaitEvents(t, events,
		expect(instanceHeartbeatOk),
		deny(instanceHeartbeatErr, instanceCompareFailed, handlerClose),
	)
	auth.mu.Lock()
	m := auth.lastInstance.GetLastMeasurement()
	auth.mu.Unlock()

	require.NotNil(t, m)
	require.InDelta(t, time.Minute, m.ControllerSystemClock.Sub(m.SystemClock)-m.RequestDuration/2, float64(time.Second))
}

type eventOpts struct {
	expect map[testEvent]int
	deny   map[testEvent]struct{}
}

type eventOption func(*eventOpts)

func expect(events ...testEvent) eventOption {
	return func(opts *eventOpts) {
		for _, event := range events {
			opts.expect[event] = opts.expect[event] + 1
		}
	}
}

func deny(events ...testEvent) eventOption {
	return func(opts *eventOpts) {
		for _, event := range events {
			opts.deny[event] = struct{}{}
		}
	}
}

func awaitEvents(t *testing.T, ch <-chan testEvent, opts ...eventOption) {
	t.Helper()
	options := eventOpts{
		expect: make(map[testEvent]int),
		deny:   make(map[testEvent]struct{}),
	}
	for _, opt := range opts {
		opt(&options)
	}

	timeout := time.After(time.Second * 30)
	for {
		if len(options.expect) == 0 {
			return
		}

		select {
		case event := <-ch:
			if _, ok := options.deny[event]; ok {
				require.Failf(t, "unexpected event", "event=%v", event)
			}

			options.expect[event] = options.expect[event] - 1
			if options.expect[event] < 1 {
				delete(options.expect, event)
			}
		case <-timeout:
			require.Failf(t, "timeout waiting for events", "expect=%+v", options.expect)
		}
	}
}

type resourceCounter struct {
	mu sync.Mutex
	c  map[string]int
}

func (r *resourceCounter) onConnect(typ string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.c == nil {
		r.c = make(map[string]int)
	}
	r.c[typ]++
}

func (r *resourceCounter) onDisconnect(typ string, amount int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.c == nil {
		r.c = make(map[string]int)
	}
	r.c[typ] -= amount
}

func (r *resourceCounter) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	var count int
	for _, v := range r.c {
		count += v
	}
	return count
}
