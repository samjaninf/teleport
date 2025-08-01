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

package rollout

import (
	"context"
	"time"

	"github.com/gravitational/trace"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gravitational/teleport/api/gen/proto/go/teleport/autoupdate/v1"
	"github.com/gravitational/teleport/api/types"
)

const (
	// Common update reasons
	updateReasonCreated         = "created"
	updateReasonReconcilerError = "reconciler_error"
	updateReasonRolloutChanged  = "rollout_changed_during_window"
)

// rolloutStrategy is responsible for rolling out the update across groups.
// This interface allows us to inject dummy strategies for simpler testing.
type rolloutStrategy interface {
	name() string
	// progressRollout takes the new rollout spec, existing rollout status and current time.
	// It updates the status resource in-place to progress the rollout to the next step if possible/needed.
	progressRollout(context.Context, *autoupdate.AutoUpdateAgentRolloutSpec, *autoupdate.AutoUpdateAgentRolloutStatus, time.Time) error
}

// inWindow checks if the time is in the group's maintenance window.
// The maintenance window is the semi-open interval: [windowStart, windowEnd).
func inWindow(group *autoupdate.AutoUpdateAgentRolloutStatusGroup, now time.Time, duration time.Duration) (bool, error) {
	dayOK, err := canUpdateToday(group.ConfigDays, now)
	if err != nil {
		return false, trace.Wrap(err, "checking the day of the week")
	}
	if !dayOK {
		return false, nil
	}

	// We compute the theoretical window start and end
	windowStart := now.Truncate(24 * time.Hour).Add(time.Duration(group.ConfigStartHour) * time.Hour)
	windowEnd := windowStart.Add(duration)

	return !now.Before(windowStart) && now.Before(windowEnd), nil
}

// rolloutChangedInWindow checks if the rollout got created after the theoretical group start time
func rolloutChangedInWindow(group *autoupdate.AutoUpdateAgentRolloutStatusGroup, now, rolloutStart time.Time, duration time.Duration) (bool, error) {
	// If the rollout is older than 24h, we know it did not change during the window
	if now.Sub(rolloutStart) > 24*time.Hour {
		return false, nil
	}
	// Else we check if the rollout happened in the group window.
	return inWindow(group, rolloutStart, duration)
}

func canUpdateToday(allowedDays []string, now time.Time) (bool, error) {
	for _, allowedDay := range allowedDays {
		if allowedDay == types.Wildcard {
			return true, nil
		}
		weekday, ok := types.ParseWeekday(allowedDay)
		if !ok {
			return false, trace.BadParameter("failed to parse weekday %q", allowedDay)
		}
		if weekday == now.Weekday() {
			return true, nil
		}
	}
	return false, nil
}

func setGroupState(group *autoupdate.AutoUpdateAgentRolloutStatusGroup, newState autoupdate.AutoUpdateAgentGroupState, reason string, now time.Time) {
	changed := false
	previousState := group.State

	// Check if there is a state transition
	if previousState != newState {
		group.State = newState
		changed = true
		// If we just started the group, also update the start time.
		// If we are doing a canary -> active transition, we don't override the start date.
		if (newState == autoupdate.AutoUpdateAgentGroupState_AUTO_UPDATE_AGENT_GROUP_STATE_ACTIVE &&
			previousState != autoupdate.AutoUpdateAgentGroupState_AUTO_UPDATE_AGENT_GROUP_STATE_CANARY) ||
			newState == autoupdate.AutoUpdateAgentGroupState_AUTO_UPDATE_AGENT_GROUP_STATE_CANARY {
			group.StartTime = timestamppb.New(now)
		}
	}

	// Check if there is a reason change. Even if the state did not change, we
	// might want to explain why.
	if group.LastUpdateReason != reason {
		group.LastUpdateReason = reason
		changed = true
	}

	if changed {
		group.LastUpdateTime = timestamppb.New(now)
	}
}

func computeRolloutState(groups []*autoupdate.AutoUpdateAgentRolloutStatusGroup) autoupdate.AutoUpdateAgentRolloutState {
	groupCount := len(groups)

	if groupCount == 0 {
		return autoupdate.AutoUpdateAgentRolloutState_AUTO_UPDATE_AGENT_ROLLOUT_STATE_UNSPECIFIED
	}

	var doneGroups, unstartedGroups int

	for _, group := range groups {
		switch group.State {
		// If one or more groups have been rolled back, we consider the rollout rolledback
		case autoupdate.AutoUpdateAgentGroupState_AUTO_UPDATE_AGENT_GROUP_STATE_ROLLEDBACK:
			return autoupdate.AutoUpdateAgentRolloutState_AUTO_UPDATE_AGENT_ROLLOUT_STATE_ROLLEDBACK

		case autoupdate.AutoUpdateAgentGroupState_AUTO_UPDATE_AGENT_GROUP_STATE_UNSTARTED:
			unstartedGroups++

		case autoupdate.AutoUpdateAgentGroupState_AUTO_UPDATE_AGENT_GROUP_STATE_DONE:
			doneGroups++
		}
	}

	// If every group is done, the rollout is done.
	if doneGroups == groupCount {
		return autoupdate.AutoUpdateAgentRolloutState_AUTO_UPDATE_AGENT_ROLLOUT_STATE_DONE
	}

	// If every group is unstarted, the rollout is unstarted.
	if unstartedGroups == groupCount {
		return autoupdate.AutoUpdateAgentRolloutState_AUTO_UPDATE_AGENT_ROLLOUT_STATE_UNSTARTED
	}

	// Else at least one group is active or done, but not everything is finished. We consider the rollout active.
	return autoupdate.AutoUpdateAgentRolloutState_AUTO_UPDATE_AGENT_ROLLOUT_STATE_ACTIVE
}
