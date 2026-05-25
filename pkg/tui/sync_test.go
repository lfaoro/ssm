// Copyright (c) 2026 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestSyncModel_NilBase(t *testing.T) {
	m := SyncModel(nil)
	if m != nil {
		t.Error("SyncModel(nil) should return nil")
	}
}

func TestSyncModel_Init(t *testing.T) {
	m := newTestModel(t, false)
	sm := SyncModel(m)
	sm2, ok := sm.(*syncModel)
	if !ok {
		t.Fatal("SyncModel() should return *syncModel")
	}
	if len(sm2.providers) != 4 {
		t.Errorf("expected 4 providers, got %d", len(sm2.providers))
	}
	if sm2.status != syncIdle {
		t.Errorf("expected idle status")
	}
	cmd := sm2.Init()
	if cmd != nil {
		t.Error("expected nil init command")
	}
}

func TestSyncModel_EscReturnsPrevious(t *testing.T) {
	m := newTestModel(t, false)
	sm := SyncModel(m)

	result, _ := sm.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if result != m {
		t.Error("expected previous model for KeyEsc")
	}
}

func TestSyncModel_QReturnsPrevious(t *testing.T) {
	m := newTestModel(t, false)
	sm := SyncModel(m)

	result, _ := sm.Update(tea.KeyPressMsg{Code: 'q'})
	if result != m {
		t.Error("expected previous model for 'q'")
	}
}

func TestSyncModel_CompleteUpdatesStatus(t *testing.T) {
	m := newTestModel(t, false)
	sm := SyncModel(m).(*syncModel) //nolint:forcetypeassert

	msg := syncCompleteMsg{
		progress: []syncProgressMsg{
			{provider: "hetzner", count: 3},
			{provider: "aws", count: 5},
		},
		summary: []string{"Hetzner: 3 servers", "AWS: 5 servers"},
	}
	result, _ := sm.Update(msg)
	sm2, ok := result.(*syncModel)
	if !ok {
		t.Fatal("expected *syncModel return")
	}
	if sm2.status != syncDone {
		t.Errorf("expected syncDone, got %v", sm2.status)
	}
	if sm2.resultLines == nil || len(sm2.resultLines) != 2 {
		t.Errorf("expected 2 result lines, got %d", len(sm2.resultLines))
	}
	for _, p := range sm2.providers {
		if p.name == "hetzner" && p.status != provDone {
			t.Errorf("expected hetzner provDone, got %v", p.status)
		}
		if p.name == "aws" && p.count != 5 {
			t.Errorf("expected aws count 5, got %d", p.count)
		}
	}
}

func TestSyncModel_CompleteNoServers(t *testing.T) {
	m := newTestModel(t, false)
	sm := SyncModel(m).(*syncModel) //nolint:forcetypeassert

	msg := syncCompleteMsg{
		progress: []syncProgressMsg{
			{provider: "hetzner", count: 0},
		},
		summary: []string{"No servers found (check credentials)"},
	}
	result, _ := sm.Update(msg)
	sm2, ok := result.(*syncModel)
	if !ok {
		t.Fatal("expected *syncModel return")
	}
	if len(sm2.resultLines) != 1 || sm2.resultLines[0] != "No servers found (check credentials)" {
		t.Errorf("unexpected summary: %v", sm2.resultLines)
	}
}

func TestSyncModel_EnterStartsSync(t *testing.T) {
	m := newTestModel(t, false)
	sm := SyncModel(m).(*syncModel) //nolint:forcetypeassert

	result, cmd := sm.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	sm2, ok := result.(*syncModel)
	if !ok {
		t.Fatal("expected *syncModel return")
	}
	if sm2.status != syncRunning {
		t.Errorf("expected syncRunning, got %v", sm2.status)
	}
	if cmd == nil {
		t.Error("expected non-nil command for sync start")
	}
}

func TestSyncModel_ProgressUpdatesProvider(t *testing.T) {
	m := newTestModel(t, false)
	sm := SyncModel(m).(*syncModel) //nolint:forcetypeassert

	msg := syncProgressMsg{provider: "gcp"}
	sm.Update(msg)
	for _, p := range sm.providers {
		if p.name == "gcp" && p.status != provFetching {
			t.Errorf("expected gcp provFetching, got %v", p.status)
		}
	}
}

func TestSyncModel_StartSyncCreatesCommand(t *testing.T) {
	m := newTestModel(t, false)
	sm := SyncModel(m).(*syncModel) //nolint:forcetypeassert

	cmd := sm.startSync()
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
}
