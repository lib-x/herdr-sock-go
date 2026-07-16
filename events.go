package herdrsock

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

type EventsSubscribeParams struct {
	Subscriptions []Subscription `json:"subscriptions"`
}

// Subscription is a flexible event subscription payload. Use the helper
// constructors below or build a map for newer Herdr event types.
type Subscription map[string]any

func SubscribeWorkspaceCreated() Subscription { return Subscription{"type": "workspace.created"} }
func SubscribeWorkspaceUpdated() Subscription { return Subscription{"type": "workspace.updated"} }
func SubscribeWorkspaceMetadataUpdated() Subscription {
	return Subscription{"type": "workspace.metadata_updated"}
}
func SubscribeWorkspaceRenamed() Subscription { return Subscription{"type": "workspace.renamed"} }
func SubscribeWorkspaceMoved() Subscription   { return Subscription{"type": "workspace.moved"} }
func SubscribeWorkspaceClosed() Subscription  { return Subscription{"type": "workspace.closed"} }
func SubscribeWorkspaceFocused() Subscription { return Subscription{"type": "workspace.focused"} }
func SubscribeWorktreeCreated() Subscription  { return Subscription{"type": "worktree.created"} }
func SubscribeWorktreeOpened() Subscription   { return Subscription{"type": "worktree.opened"} }
func SubscribeWorktreeRemoved() Subscription  { return Subscription{"type": "worktree.removed"} }
func SubscribeTabCreated() Subscription       { return Subscription{"type": "tab.created"} }
func SubscribeTabClosed() Subscription        { return Subscription{"type": "tab.closed"} }
func SubscribeTabFocused() Subscription       { return Subscription{"type": "tab.focused"} }
func SubscribeTabRenamed() Subscription       { return Subscription{"type": "tab.renamed"} }
func SubscribeTabMoved() Subscription         { return Subscription{"type": "tab.moved"} }
func SubscribePaneCreated() Subscription      { return Subscription{"type": "pane.created"} }
func SubscribePaneClosed() Subscription       { return Subscription{"type": "pane.closed"} }
func SubscribePaneUpdated() Subscription      { return Subscription{"type": "pane.updated"} }
func SubscribePaneFocused() Subscription      { return Subscription{"type": "pane.focused"} }
func SubscribePaneMoved() Subscription        { return Subscription{"type": "pane.moved"} }
func SubscribePaneExited() Subscription       { return Subscription{"type": "pane.exited"} }
func SubscribePaneAgentDetected() Subscription {
	return Subscription{"type": "pane.agent_detected"}
}

func SubscribeLayoutUpdated() Subscription { return Subscription{"type": "layout.updated"} }

func SubscribePaneAgentStatusChanged(paneID string, status *AgentStatus) Subscription {
	sub := Subscription{"type": "pane.agent_status_changed", "pane_id": paneID}
	if status != nil {
		sub["agent_status"] = *status
	}
	return sub
}

func SubscribePaneOutputMatched(paneID string, source ReadSource, match OutputMatch, lines *uint32, stripANSI bool) Subscription {
	sub := Subscription{
		"type":       "pane.output_matched",
		"pane_id":    paneID,
		"source":     source,
		"match":      match,
		"strip_ansi": stripANSI,
	}
	if lines != nil {
		sub["lines"] = *lines
	}
	return sub
}

func SubscribePaneScrollChanged(paneID string) Subscription {
	return Subscription{"type": "pane.scroll_changed", "pane_id": paneID}
}

type OutputMatch map[string]string

func SubstringMatch(value string) OutputMatch {
	return OutputMatch{"type": "substring", "value": value}
}

func RegexMatch(value string) OutputMatch {
	return OutputMatch{"type": "regex", "value": value}
}

// EventStream is a long-lived events.subscribe connection.
type EventStream struct {
	conn   net.Conn
	reader *bufio.Reader
}

// SubscriptionEvent is the raw pushed event envelope.
type SubscriptionEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type subscriptionStartedResult struct {
	Type string `json:"type"`
}

// Subscribe starts an events.subscribe request and returns the acknowledged
// stream. Close the stream when done.
func (c *Client) Subscribe(ctx context.Context, subscriptions ...Subscription) (*EventStream, error) {
	return c.SubscribeWithID(ctx, "", subscriptions...)
}

func (c *Client) SubscribeWithID(ctx context.Context, id string, subscriptions ...Subscription) (*EventStream, error) {
	if id == "" {
		id = nextRequestID()
	}
	conn, err := c.dial(ctx)
	if err != nil {
		return nil, err
	}

	if deadline, ok := c.deadline(ctx); ok {
		_ = conn.SetDeadline(deadline)
	}

	request := Request{
		ID:     id,
		Method: MethodEventsSubscribe,
		Params: EventsSubscribeParams{Subscriptions: subscriptions},
	}
	if err := writeJSONLine(conn, request); err != nil {
		_ = conn.Close()
		return nil, err
	}

	reader := bufio.NewReader(conn)
	var response Response
	if err := readJSONLine(reader, &response); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if response.ID != id {
		_ = conn.Close()
		return nil, &IDMismatchError{Want: id, Got: response.ID}
	}
	if response.Error != nil {
		_ = conn.Close()
		return nil, &ErrorResponse{ID: response.ID, Body: *response.Error}
	}
	var ack subscriptionStartedResult
	if err := json.Unmarshal(response.Result, &ack); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("decode events.subscribe ack: %w", err)
	}
	if ack.Type != "subscription_started" {
		_ = conn.Close()
		return nil, fmt.Errorf("events.subscribe returned %q, want subscription_started", ack.Type)
	}

	_ = conn.SetDeadline(time.Time{})
	return &EventStream{conn: conn, reader: reader}, nil
}

func (s *EventStream) Close() error {
	if s == nil || s.conn == nil {
		return nil
	}
	return s.conn.Close()
}

func (s *EventStream) Next(ctx context.Context) (*SubscriptionEvent, error) {
	if s == nil || s.conn == nil || s.reader == nil {
		return nil, fmt.Errorf("herdrsock: event stream is closed")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if ctx != nil {
		if deadline, ok := ctx.Deadline(); ok {
			_ = s.conn.SetReadDeadline(deadline)
		} else {
			_ = s.conn.SetReadDeadline(time.Time{})
		}
	}
	if done := ctx.Done(); done != nil {
		cancelled := make(chan struct{})
		go func() {
			select {
			case <-done:
				_ = s.conn.SetReadDeadline(time.Now())
			case <-cancelled:
			}
		}()
		defer close(cancelled)
	}
	defer func() {
		_ = s.conn.SetReadDeadline(time.Time{})
	}()

	var event SubscriptionEvent
	if err := readJSONLine(s.reader, &event); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil && (isTimeout(err) || errors.Is(err, os.ErrDeadlineExceeded)) {
			return nil, ctxErr
		}
		return nil, err
	}
	return &event, nil
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func (s *EventStream) DecodeData(event *SubscriptionEvent, out any) error {
	if event == nil {
		return fmt.Errorf("herdrsock: nil event")
	}
	return json.Unmarshal(event.Data, out)
}

type PaneAgentStatusChangedEvent struct {
	PaneID      string      `json:"pane_id"`
	WorkspaceID string      `json:"workspace_id"`
	AgentStatus AgentStatus `json:"agent_status"`
	Agent       *string     `json:"agent,omitempty"`
	// Deprecated: Herdr 0.7.4 no longer emits custom_status. Use AgentStatus,
	// Title, DisplayAgent, and StateLabels instead.
	CustomStatus *string           `json:"custom_status,omitempty"`
	Title        *string           `json:"title,omitempty"`
	DisplayAgent *string           `json:"display_agent,omitempty"`
	StateLabels  map[string]string `json:"state_labels,omitempty"`
}

type PaneOutputMatchedEvent struct {
	PaneID      string         `json:"pane_id"`
	MatchedLine string         `json:"matched_line"`
	Read        PaneReadResult `json:"read"`
}

type PaneScrollChangedEvent struct {
	PaneID      string         `json:"pane_id"`
	WorkspaceID string         `json:"workspace_id"`
	Scroll      PaneScrollInfo `json:"scroll"`
}

type WorkspaceCreatedEvent struct {
	Type      string        `json:"type"`
	Workspace WorkspaceInfo `json:"workspace"`
}

type WorkspaceUpdatedEvent struct {
	Type      string        `json:"type"`
	Workspace WorkspaceInfo `json:"workspace"`
}

type WorkspaceMetadataUpdatedEvent struct {
	Type      string        `json:"type"`
	Workspace WorkspaceInfo `json:"workspace"`
}

type WorkspaceRenamedEvent struct {
	Type        string `json:"type"`
	WorkspaceID string `json:"workspace_id"`
	Label       string `json:"label"`
}

type WorkspaceMovedEvent struct {
	Type        string          `json:"type"`
	WorkspaceID string          `json:"workspace_id"`
	InsertIndex int             `json:"insert_index"`
	Workspaces  []WorkspaceInfo `json:"workspaces"`
}

type WorkspaceClosedEvent struct {
	Type        string         `json:"type"`
	WorkspaceID string         `json:"workspace_id"`
	Workspace   *WorkspaceInfo `json:"workspace,omitempty"`
}

type WorkspaceFocusedEvent struct {
	Type        string `json:"type"`
	WorkspaceID string `json:"workspace_id"`
}

type WorktreeCreatedEvent struct {
	Type      string        `json:"type"`
	Workspace WorkspaceInfo `json:"workspace"`
	Worktree  WorktreeInfo  `json:"worktree"`
}

type WorktreeOpenedEvent struct {
	Type        string        `json:"type"`
	Workspace   WorkspaceInfo `json:"workspace"`
	Worktree    WorktreeInfo  `json:"worktree"`
	AlreadyOpen bool          `json:"already_open"`
}

type WorktreeRemovedEvent struct {
	Type        string         `json:"type"`
	WorkspaceID string         `json:"workspace_id"`
	Workspace   *WorkspaceInfo `json:"workspace,omitempty"`
	Worktree    WorktreeInfo   `json:"worktree"`
	Forced      bool           `json:"forced"`
}

type TabCreatedEvent struct {
	Type string  `json:"type"`
	Tab  TabInfo `json:"tab"`
}

type TabClosedEvent struct {
	Type        string `json:"type"`
	TabID       string `json:"tab_id"`
	WorkspaceID string `json:"workspace_id"`
}

type TabFocusedEvent struct {
	Type        string `json:"type"`
	TabID       string `json:"tab_id"`
	WorkspaceID string `json:"workspace_id"`
}

type TabRenamedEvent struct {
	Type        string `json:"type"`
	TabID       string `json:"tab_id"`
	WorkspaceID string `json:"workspace_id"`
	Label       string `json:"label"`
}

type TabMovedEvent struct {
	Type        string    `json:"type"`
	TabID       string    `json:"tab_id"`
	WorkspaceID string    `json:"workspace_id"`
	InsertIndex int       `json:"insert_index"`
	Tabs        []TabInfo `json:"tabs"`
}

type PaneUpdatedEvent struct {
	Type string   `json:"type"`
	Pane PaneInfo `json:"pane"`
}

type LayoutUpdatedEvent struct {
	Type   string             `json:"type"`
	Layout PaneLayoutSnapshot `json:"layout"`
}
