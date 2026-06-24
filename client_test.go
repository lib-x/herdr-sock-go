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
		writeLine(t, w, `{"id":"req_1","result":{"type":"pong","version":"0.7.0","protocol":14,"capabilities":{"live_handoff":true}}}`)
	})

	client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
	var pong PongResult
	if err := client.Call(context.Background(), "req_1", MethodPing, nil, &pong); err != nil {
		t.Fatal(err)
	}
	if pong.Protocol != CurrentProtocol || pong.Version != "0.7.0" {
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
