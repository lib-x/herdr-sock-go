package herdrsock

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestSocketPathResolution(t *testing.T) {
	t.Setenv(SocketPathEnv, "")
	t.Setenv(SessionEnv, "")
	t.Setenv("XDG_CONFIG_HOME", "/tmp/herdrsock-config")

	path, err := ActiveSocketPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/tmp/herdrsock-config", "herdr", "herdr.sock")
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}

	t.Setenv(SessionEnv, "work")
	path, err = ActiveSocketPath()
	if err != nil {
		t.Fatal(err)
	}
	want = filepath.Join("/tmp/herdrsock-config", "herdr", "sessions", "work", "herdr.sock")
	if path != want {
		t.Fatalf("session path = %q, want %q", path, want)
	}

	t.Setenv(SocketPathEnv, "/tmp/custom-herdr.sock")
	path, err = ActiveSocketPath()
	if err != nil {
		t.Fatal(err)
	}
	if path != "/tmp/custom-herdr.sock" {
		t.Fatalf("override path = %q", path)
	}
}

func TestValidateSessionName(t *testing.T) {
	valid := []string{"work", "a.b_c-1", strings.Repeat("a", 64)}
	for _, name := range valid {
		if err := ValidateSessionName(name); err != nil {
			t.Fatalf("ValidateSessionName(%q): %v", name, err)
		}
	}

	invalid := []string{"", ".", "..", "has/slash", "white space", strings.Repeat("a", 65)}
	for _, name := range invalid {
		if err := ValidateSessionName(name); err == nil {
			t.Fatalf("ValidateSessionName(%q) succeeded", name)
		}
	}
}

func TestCallRawSuccess(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket test")
	}
	socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
		if request["method"] != MethodPing {
			t.Fatalf("method = %v", request["method"])
		}
		params, ok := request["params"].(map[string]any)
		if !ok || len(params) != 0 {
			t.Fatalf("params = %#v, want empty object", request["params"])
		}
		writeLine(t, w, `{"id":"req_1","result":{"type":"pong","version":"0.7.4","protocol":16,"capabilities":{"live_handoff":true,"detached_server_daemon":true}}}`)
	})

	client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
	var pong PongResult
	if err := client.Call(context.Background(), "req_1", MethodPing, nil, &pong); err != nil {
		t.Fatal(err)
	}
	if pong.Protocol != CurrentProtocol || pong.Version != CurrentHerdrVersion || pong.Capabilities == nil || !pong.Capabilities.DetachedServerDaemon {
		t.Fatalf("pong = %#v", pong)
	}
}

func TestCallRawErrorResponse(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket test")
	}
	socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
		writeLine(t, w, `{"id":"bad","error":{"code":"not_found","message":"pane not found"}}`)
	})

	client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
	_, err := client.CallRaw(context.Background(), "bad", MethodPaneGet, PaneTarget{PaneID: "w1:p9"})
	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T %[1]v, want ErrorResponse", err)
	}
	if apiErr.Body.Code != "not_found" {
		t.Fatalf("code = %q", apiErr.Body.Code)
	}
}

func TestReadPaneOmitsStripANSIByDefault(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket test")
	}
	socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
		id, _ := request["id"].(string)
		params, ok := request["params"].(map[string]any)
		if !ok {
			t.Fatalf("params = %#v", request["params"])
		}
		if _, ok := params["strip_ansi"]; ok {
			t.Fatalf("strip_ansi was sent by default: %#v", params)
		}
		writeLine(t, w, `{"id":"`+id+`","result":{"type":"pane_read","read":{"pane_id":"w1:p1","workspace_id":"w1","tab_id":"w1:t1","source":"recent","format":"text","text":"ok","revision":1,"truncated":false}}}`)
	})

	client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
	read, err := client.ReadPane(context.Background(), PaneReadParams{
		PaneID: "w1:p1",
		Source: ReadRecent,
	})
	if err != nil {
		t.Fatal(err)
	}
	if read.Text != "ok" {
		t.Fatalf("read text = %q", read.Text)
	}
}

func TestRequireProtocolMismatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket test")
	}
	socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
		id, _ := request["id"].(string)
		writeLine(t, w, `{"id":"`+id+`","result":{"type":"pong","version":"0.6.0","protocol":13}}`)
	})

	client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
	_, err := client.RequireCurrentProtocol(context.Background())
	var mismatch *ProtocolMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("err = %T %[1]v, want ProtocolMismatchError", err)
	}
	if mismatch.Required != CurrentProtocol || mismatch.Actual != 13 {
		t.Fatalf("mismatch = %#v", mismatch)
	}
}

func TestSubscribeReadsEvents(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket test")
	}
	socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
		if request["method"] != MethodEventsSubscribe {
			t.Fatalf("method = %v", request["method"])
		}
		writeLine(t, w, `{"id":"sub","result":{"type":"subscription_started"}}`)
		writeLine(t, w, `{"event":"pane.agent_status_changed","data":{"pane_id":"w1:p1","workspace_id":"w1","agent_status":"done"}}`)
	})

	client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
	stream, err := client.SubscribeWithID(context.Background(), "sub", SubscribePaneAgentStatusChanged("w1:p1", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	event, err := stream.Next(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if event.Event != "pane.agent_status_changed" {
		t.Fatalf("event = %q", event.Event)
	}
	var data PaneAgentStatusChangedEvent
	if err := stream.DecodeData(event, &data); err != nil {
		t.Fatal(err)
	}
	if data.PaneID != "w1:p1" || data.AgentStatus != AgentStatusDone {
		t.Fatalf("data = %#v", data)
	}
}

func TestWorktreeSubscriptionHelpers(t *testing.T) {
	tests := []struct {
		name string
		sub  Subscription
		want string
	}{
		{name: "workspace moved", sub: SubscribeWorkspaceMoved(), want: "workspace.moved"},
		{name: "created", sub: SubscribeWorktreeCreated(), want: "worktree.created"},
		{name: "opened", sub: SubscribeWorktreeOpened(), want: "worktree.opened"},
		{name: "removed", sub: SubscribeWorktreeRemoved(), want: "worktree.removed"},
		{name: "tab created", sub: SubscribeTabCreated(), want: "tab.created"},
		{name: "tab closed", sub: SubscribeTabClosed(), want: "tab.closed"},
		{name: "tab focused", sub: SubscribeTabFocused(), want: "tab.focused"},
		{name: "tab renamed", sub: SubscribeTabRenamed(), want: "tab.renamed"},
		{name: "tab moved", sub: SubscribeTabMoved(), want: "tab.moved"},
		{name: "layout updated", sub: SubscribeLayoutUpdated(), want: "layout.updated"},
		{name: "pane scroll changed", sub: SubscribePaneScrollChanged("w1:p1"), want: "pane.scroll_changed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.sub["type"].(string)
			if !ok || got != tt.want {
				t.Fatalf("subscription type = %#v, want %q", tt.sub["type"], tt.want)
			}
		})
	}
}

func TestSessionSnapshot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket test")
	}
	socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
		if request["method"] != MethodSessionSnapshot {
			t.Fatalf("method = %v", request["method"])
		}
		id, _ := request["id"].(string)
		writeLine(t, w, `{"id":"`+id+`","result":{"type":"session_snapshot","snapshot":{"version":"0.7.4","protocol":16,"focused_workspace_id":"w1","focused_tab_id":"w1:t1","focused_pane_id":"w1:p1","workspaces":[],"tabs":[],"panes":[],"layouts":[],"agents":[]}}}`)
	})

	client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
	snapshot, err := client.SessionSnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Version != CurrentHerdrVersion || snapshot.Protocol != CurrentProtocol || snapshot.FocusedPaneID == nil || *snapshot.FocusedPaneID != "w1:p1" {
		t.Fatalf("snapshot = %#v", snapshot)
	}
}

func TestDecodeV074PresentationMetadata(t *testing.T) {
	line := []byte(`{
		"version":"0.7.4",
		"protocol":16,
		"workspaces":[{
			"workspace_id":"w1",
			"number":1,
			"label":"repo",
			"focused":true,
			"pane_count":1,
			"tab_count":1,
			"active_tab_id":"w1:t1",
			"agent_status":"working",
			"tokens":{"summary":"review ready"}
		}],
		"tabs":[],
		"panes":[{
			"pane_id":"w1:p1",
			"terminal_id":"term-1",
			"workspace_id":"w1",
			"tab_id":"w1:t1",
			"focused":true,
			"terminal_title":"⠋ Codex",
			"terminal_title_stripped":"Codex",
			"agent_status":"working",
			"tokens":{"model":"gpt-5"},
			"revision":3
		}],
		"layouts":[],
		"agents":[{
			"terminal_id":"term-1",
			"workspace_id":"w1",
			"tab_id":"w1:t1",
			"pane_id":"w1:p1",
			"focused":true,
			"agent_status":"working",
			"terminal_title":"⠋ Codex",
			"terminal_title_stripped":"Codex",
			"tokens":{"model":"gpt-5"},
			"revision":3
		}]
	}`)

	var snapshot SessionSnapshot
	if err := json.Unmarshal(line, &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.Workspaces[0].Tokens["summary"] != "review ready" {
		t.Fatalf("workspace = %#v", snapshot.Workspaces[0])
	}
	if snapshot.Panes[0].TerminalTitleStripped == nil || *snapshot.Panes[0].TerminalTitleStripped != "Codex" || snapshot.Panes[0].Tokens["model"] != "gpt-5" {
		t.Fatalf("pane = %#v", snapshot.Panes[0])
	}
	if snapshot.Agents[0].TerminalTitle == nil || *snapshot.Agents[0].TerminalTitle != "⠋ Codex" || snapshot.Agents[0].Tokens["model"] != "gpt-5" {
		t.Fatalf("agent = %#v", snapshot.Agents[0])
	}
}

func TestCreateWorktreeSendsParams(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket test")
	}
	socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
		if request["method"] != MethodWorktreeCreate {
			t.Fatalf("method = %v", request["method"])
		}
		params, ok := request["params"].(map[string]any)
		if !ok {
			t.Fatalf("params = %#v", request["params"])
		}
		if params["branch"] != "feature/api" || params["focus"] != true {
			t.Fatalf("params = %#v", params)
		}
		id, _ := request["id"].(string)
		writeLine(t, w, `{"id":"`+id+`","result":{"type":"worktree_created","workspace":{"workspace_id":"w2","number":2,"label":"feature","focused":true,"pane_count":1,"tab_count":1,"active_tab_id":"w2:t1","agent_status":"unknown"},"tab":{"tab_id":"w2:t1","workspace_id":"w2","number":1,"label":"1","focused":true,"pane_count":1,"agent_status":"unknown"},"root_pane":{"pane_id":"w2:p1","terminal_id":"term2","workspace_id":"w2","tab_id":"w2:t1","focused":true,"agent_status":"unknown","revision":1},"worktree":{"path":"/repo/herdr-feature","branch":"feature/api","is_bare":false,"is_detached":false,"is_prunable":false,"is_linked_worktree":true,"label":"herdr"}}}`)
	})

	branch := "feature/api"
	client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
	workspace, _, _, worktree, err := client.CreateWorktree(context.Background(), WorktreeCreateParams{
		Branch: &branch,
		Focus:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if workspace.WorkspaceID != "w2" || worktree.Branch == nil || *worktree.Branch != branch {
		t.Fatalf("workspace = %#v, worktree = %#v", workspace, worktree)
	}
}

func TestSetSplitRatio(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket test")
	}
	socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
		if request["method"] != MethodLayoutSetSplitRatio {
			t.Fatalf("method = %v", request["method"])
		}
		params, ok := request["params"].(map[string]any)
		if !ok {
			t.Fatalf("params = %#v", request["params"])
		}
		if params["ratio"] != 0.6 {
			t.Fatalf("ratio = %#v", params["ratio"])
		}
		id, _ := request["id"].(string)
		writeLine(t, w, `{"id":"`+id+`","result":{"type":"layout_split_ratio_set","layout":{"workspace_id":"w1","tab_id":"w1:t1","zoomed":false,"focused_pane_id":"w1:p1","root":{"type":"split","direction":"right","ratio":0.6,"first":{"type":"pane","pane_id":"w1:p1"},"second":{"type":"pane","pane_id":"w1:p2"}}}}}`)
	})

	client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
	layout, err := client.SetSplitRatio(context.Background(), LayoutSetSplitRatioParams{
		TabID: stringPtr("w1:t1"),
		Path:  []bool{},
		Ratio: 0.6,
	})
	if err != nil {
		t.Fatal(err)
	}
	if layout.Root.Ratio == nil || *layout.Root.Ratio != 0.6 {
		t.Fatalf("layout = %#v", layout)
	}
}

func TestDecodeWorktreeRemovedEventWithWorkspaceSnapshot(t *testing.T) {
	line := []byte(`{
		"event":"worktree_removed",
		"data":{
			"type":"worktree_removed",
			"workspace_id":"w2",
			"workspace":{
				"workspace_id":"w2",
				"number":2,
				"label":"feature",
				"focused":false,
				"pane_count":1,
				"tab_count":1,
				"active_tab_id":"w2:t1",
				"agent_status":"unknown"
			},
			"worktree":{
				"path":"/repo/herdr-feature",
				"branch":"feature/api",
				"is_bare":false,
				"is_detached":false,
				"is_prunable":false,
				"is_linked_worktree":true,
				"label":"herdr"
			},
			"forced":true
		}
	}`)

	var event SubscriptionEvent
	if err := json.Unmarshal(line, &event); err != nil {
		t.Fatal(err)
	}
	var data WorktreeRemovedEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		t.Fatal(err)
	}

	if event.Event != "worktree_removed" {
		t.Fatalf("event = %q", event.Event)
	}
	if data.Type != "worktree_removed" || data.WorkspaceID != "w2" || !data.Forced {
		t.Fatalf("data = %#v", data)
	}
	if data.Workspace == nil || data.Workspace.WorkspaceID != "w2" {
		t.Fatalf("workspace snapshot = %#v", data.Workspace)
	}
	if data.Worktree.Branch == nil || *data.Worktree.Branch != "feature/api" {
		t.Fatalf("worktree = %#v", data.Worktree)
	}
}

func TestDecodePaneScrollAndLayoutEvents(t *testing.T) {
	scrollLine := []byte(`{
		"event":"pane.scroll_changed",
		"data":{
			"pane_id":"w1:p1",
			"workspace_id":"w1",
			"scroll":{
				"offset_from_bottom":12,
				"max_offset_from_bottom":240,
				"viewport_rows":30
			}
		}
	}`)
	var scrollEvent SubscriptionEvent
	if err := json.Unmarshal(scrollLine, &scrollEvent); err != nil {
		t.Fatal(err)
	}
	var scrollData PaneScrollChangedEvent
	if err := json.Unmarshal(scrollEvent.Data, &scrollData); err != nil {
		t.Fatal(err)
	}
	if scrollData.Scroll.OffsetFromBottom != 12 || scrollData.Scroll.ViewportRows != 30 {
		t.Fatalf("scroll data = %#v", scrollData)
	}

	layoutLine := []byte(`{
		"event":"layout_updated",
		"data":{
			"type":"layout_updated",
			"layout":{
				"workspace_id":"w1",
				"tab_id":"w1:t1",
				"zoomed":false,
				"area":{"x":0,"y":0,"width":100,"height":24},
				"focused_pane_id":"w1:p1",
				"panes":[{"pane_id":"w1:p1","focused":true,"rect":{"x":0,"y":0,"width":100,"height":24}}],
				"splits":[]
			}
		}
	}`)
	var layoutEvent SubscriptionEvent
	if err := json.Unmarshal(layoutLine, &layoutEvent); err != nil {
		t.Fatal(err)
	}
	var layoutData LayoutUpdatedEvent
	if err := json.Unmarshal(layoutEvent.Data, &layoutData); err != nil {
		t.Fatal(err)
	}
	if layoutData.Type != "layout_updated" || layoutData.Layout.TabID != "w1:t1" {
		t.Fatalf("layout data = %#v", layoutData)
	}
}

func stringPtr(value string) *string {
	return &value
}

func startUnixJSONServer(t *testing.T, handle func(*testing.T, map[string]any, *bufio.Writer)) string {
	t.Helper()
	dir := t.TempDir()
	socket := filepath.Join(dir, "herdr.sock")
	ln, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = ln.Close()
		_ = os.Remove(socket)
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		line, err := reader.ReadBytes('\n')
		if err != nil {
			t.Errorf("read request: %v", err)
			return
		}
		var request map[string]any
		if err := json.Unmarshal(line, &request); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		writer := bufio.NewWriter(conn)
		handle(t, request, writer)
		_ = writer.Flush()
	}()
	t.Cleanup(func() {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("server goroutine did not exit")
		}
	})

	return socket
}

func writeLine(t *testing.T, w *bufio.Writer, line string) {
	t.Helper()
	if _, err := w.WriteString(line + "\n"); err != nil {
		t.Fatalf("write response: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush response: %v", err)
	}
}
