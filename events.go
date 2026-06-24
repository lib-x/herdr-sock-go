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
func SubscribeWorkspaceRenamed() Subscription { return Subscription{"type": "workspace.renamed"} }
func SubscribeWorkspaceClosed() Subscription  { return Subscription{"type": "workspace.closed"} }
func SubscribeWorkspaceFocused() Subscription { return Subscription{"type": "workspace.focused"} }
func SubscribePaneCreated() Subscription      { return Subscription{"type": "pane.created"} }
func SubscribePaneClosed() Subscription       { return Subscription{"type": "pane.closed"} }
func SubscribePaneFocused() Subscription      { return Subscription{"type": "pane.focused"} }
func SubscribePaneMoved() Subscription        { return Subscription{"type": "pane.moved"} }
func SubscribePaneExited() Subscription       { return Subscription{"type": "pane.exited"} }
func SubscribePaneAgentDetected() Subscription {
	return Subscription{"type": "pane.agent_detected"}
}

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
	PaneID       string            `json:"pane_id"`
	WorkspaceID  string            `json:"workspace_id"`
	AgentStatus  AgentStatus       `json:"agent_status"`
	Agent        *string           `json:"agent,omitempty"`
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
