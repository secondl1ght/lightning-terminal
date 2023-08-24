package firewalldb

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestActionStorage tests that the ActionsDB CRUD logic.
func TestActionStorage(t *testing.T) {
	tmpDir := t.TempDir()

	db, err := NewDB(tmpDir, "test.db", nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	sessionID1 := [4]byte{1, 1, 1, 1}
	action1 := &Action{
		SessionID:          sessionID1,
		ActorName:          "Autopilot",
		FeatureName:        "auto-fees",
		Trigger:            "fee too low",
		Intent:             "increase fee",
		StructuredJsonData: "{\"something\":\"nothing\"}",
		RPCMethod:          "UpdateChanPolicy",
		RPCParamsJson:      []byte("new fee"),
		AttemptedAt:        time.Unix(32100, 0),
		State:              ActionStateDone,
	}

	sessionID2 := [4]byte{2, 2, 2, 2}
	action2 := &Action{
		SessionID:     sessionID2,
		ActorName:     "Autopilot",
		FeatureName:   "rebalancer",
		Trigger:       "channels not balanced",
		Intent:        "balance",
		RPCMethod:     "SendToRoute",
		RPCParamsJson: []byte("hops, amount"),
		AttemptedAt:   time.Unix(12300, 0),
		State:         ActionStateInit,
	}

	actionsStateFilterFn := func(state ActionState) ListActionsFilterFn {
		return func(a *Action, _ bool) (bool, bool) {
			return a.State == state, true
		}
	}

	actions, _, _, err := db.ListSessionActions(
		sessionID1, actionsStateFilterFn(ActionStateDone), nil,
	)
	require.NoError(t, err)
	require.Len(t, actions, 0)

	actions, _, _, err = db.ListSessionActions(
		sessionID2, actionsStateFilterFn(ActionStateDone), nil,
	)
	require.NoError(t, err)
	require.Len(t, actions, 0)

	id, err := db.AddAction(sessionID1, action1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), id)

	id, err = db.AddAction(sessionID2, action2)
	require.NoError(t, err)
	require.Equal(t, uint64(1), id)

	actions, _, _, err = db.ListSessionActions(
		sessionID1, actionsStateFilterFn(ActionStateDone), nil,
	)
	require.NoError(t, err)
	require.Len(t, actions, 1)
	require.Equal(t, action1, actions[0])

	actions, _, _, err = db.ListSessionActions(
		sessionID2, actionsStateFilterFn(ActionStateDone), nil,
	)
	require.NoError(t, err)
	require.Len(t, actions, 0)

	err = db.SetActionState(
		&ActionLocator{
			SessionID: sessionID2,
			ActionID:  uint64(1),
		}, ActionStateDone, "",
	)
	require.NoError(t, err)

	actions, _, _, err = db.ListSessionActions(
		sessionID2, actionsStateFilterFn(ActionStateDone), nil,
	)
	require.NoError(t, err)
	require.Len(t, actions, 1)
	action2.State = ActionStateDone
	require.Equal(t, action2, actions[0])

	id, err = db.AddAction(sessionID1, action1)
	require.NoError(t, err)
	require.Equal(t, uint64(2), id)

	// Check that providing no session id and no filter function returns
	// all the actions.
	actions, _, _, err = db.ListActions(nil, &ListActionsQuery{
		IndexOffset: 0,
		MaxNum:      100,
		Reversed:    false,
	})
	require.NoError(t, err)
	require.Len(t, actions, 3)

	// Try set an error reason for a non Errored state.
	err = db.SetActionState(
		&ActionLocator{
			SessionID: sessionID2,
			ActionID:  uint64(1),
		}, ActionStateDone, "hello",
	)
	require.Error(t, err)

	// Now try move the action to errored with a reason.
	err = db.SetActionState(
		&ActionLocator{
			SessionID: sessionID2,
			ActionID:  uint64(1),
		}, ActionStateError, "fail whale",
	)
	require.NoError(t, err)

	actions, _, _, err = db.ListSessionActions(
		sessionID2, actionsStateFilterFn(ActionStateError), nil,
	)
	require.NoError(t, err)
	require.Len(t, actions, 1)
	action2.State = ActionStateError
	action2.ErrorReason = "fail whale"
	require.Equal(t, action2, actions[0])
}

// TestListActions tests some ListAction options.
// TODO(elle): cover more test cases here.
func TestListActions(t *testing.T) {
	tmpDir := t.TempDir()

	db, err := NewDB(tmpDir, "test.db", nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	sessionID1 := [4]byte{1, 1, 1, 1}
	sessionID2 := [4]byte{2, 2, 2, 2}

	actionIds := 0
	addAction := func(sessionID [4]byte) {
		actionIds++
		action := &Action{
			SessionID:          sessionID,
			ActorName:          "Autopilot",
			FeatureName:        fmt.Sprintf("%d", actionIds),
			Trigger:            "fee too low",
			Intent:             "increase fee",
			StructuredJsonData: "{\"something\":\"nothing\"}",
			RPCMethod:          "UpdateChanPolicy",
			RPCParamsJson:      []byte("new fee"),
			AttemptedAt:        time.Unix(32100, 0),
			State:              ActionStateDone,
		}

		_, err := db.AddAction(sessionID, action)
		require.NoError(t, err)
	}

	type action struct {
		sessionID [4]byte
		actionID  string
	}

	assertActions := func(dbActions []*Action, al []*action) {
		require.Len(t, dbActions, len(al))
		for i, a := range al {
			require.EqualValues(
				t, a.sessionID, dbActions[i].SessionID,
			)
			require.Equal(t, a.actionID, dbActions[i].FeatureName)
		}
	}

	addAction(sessionID1)
	addAction(sessionID1)
	addAction(sessionID1)
	addAction(sessionID1)
	addAction(sessionID2)

	actions, lastIndex, totalCount, err := db.ListActions(nil, nil)
	require.NoError(t, err)
	require.Len(t, actions, 5)
	require.EqualValues(t, 5, lastIndex)
	require.EqualValues(t, 0, totalCount)
	assertActions(actions, []*action{
		{sessionID1, "1"},
		{sessionID1, "2"},
		{sessionID1, "3"},
		{sessionID1, "4"},
		{sessionID2, "5"},
	})

	query := &ListActionsQuery{
		Reversed: true,
	}

	actions, lastIndex, totalCount, err = db.ListActions(nil, query)
	require.NoError(t, err)
	require.Len(t, actions, 5)
	require.EqualValues(t, 1, lastIndex)
	require.EqualValues(t, 0, totalCount)
	assertActions(actions, []*action{
		{sessionID2, "5"},
		{sessionID1, "4"},
		{sessionID1, "3"},
		{sessionID1, "2"},
		{sessionID1, "1"},
	})

	actions, lastIndex, totalCount, err = db.ListActions(
		nil, &ListActionsQuery{
			CountAll: true,
		},
	)
	require.NoError(t, err)
	require.Len(t, actions, 5)
	require.EqualValues(t, 5, lastIndex)
	require.EqualValues(t, 5, totalCount)
	assertActions(actions, []*action{
		{sessionID1, "1"},
		{sessionID1, "2"},
		{sessionID1, "3"},
		{sessionID1, "4"},
		{sessionID2, "5"},
	})

	actions, lastIndex, totalCount, err = db.ListActions(
		nil, &ListActionsQuery{
			CountAll: true,
			Reversed: true,
		},
	)
	require.NoError(t, err)
	require.Len(t, actions, 5)
	require.EqualValues(t, 1, lastIndex)
	require.EqualValues(t, 5, totalCount)
	assertActions(actions, []*action{
		{sessionID2, "5"},
		{sessionID1, "4"},
		{sessionID1, "3"},
		{sessionID1, "2"},
		{sessionID1, "1"},
	})

	addAction(sessionID2)
	addAction(sessionID2)
	addAction(sessionID1)
	addAction(sessionID1)
	addAction(sessionID2)

	actions, lastIndex, totalCount, err = db.ListActions(nil, nil)
	require.NoError(t, err)
	require.Len(t, actions, 10)
	require.EqualValues(t, 10, lastIndex)
	require.EqualValues(t, 0, totalCount)
	assertActions(actions, []*action{
		{sessionID1, "1"},
		{sessionID1, "2"},
		{sessionID1, "3"},
		{sessionID1, "4"},
		{sessionID2, "5"},
		{sessionID2, "6"},
		{sessionID2, "7"},
		{sessionID1, "8"},
		{sessionID1, "9"},
		{sessionID2, "10"},
	})

	actions, lastIndex, totalCount, err = db.ListActions(
		nil, &ListActionsQuery{
			MaxNum:   3,
			CountAll: true,
		},
	)
	require.NoError(t, err)
	require.Len(t, actions, 3)
	require.EqualValues(t, 3, lastIndex)
	require.EqualValues(t, 10, totalCount)
	assertActions(actions, []*action{
		{sessionID1, "1"},
		{sessionID1, "2"},
		{sessionID1, "3"},
	})

	actions, lastIndex, totalCount, err = db.ListActions(
		nil, &ListActionsQuery{
			MaxNum:      3,
			IndexOffset: 3,
		},
	)
	require.NoError(t, err)
	require.Len(t, actions, 3)
	require.EqualValues(t, 6, lastIndex)
	require.EqualValues(t, 0, totalCount)
	assertActions(actions, []*action{
		{sessionID1, "4"},
		{sessionID2, "5"},
		{sessionID2, "6"},
	})

	actions, lastIndex, totalCount, err = db.ListActions(
		nil, &ListActionsQuery{
			MaxNum:      3,
			IndexOffset: 3,
			CountAll:    true,
		},
	)
	require.NoError(t, err)
	require.Len(t, actions, 3)
	require.EqualValues(t, 6, lastIndex)
	require.EqualValues(t, 10, totalCount)
	assertActions(actions, []*action{
		{sessionID1, "4"},
		{sessionID2, "5"},
		{sessionID2, "6"},
	})
}
