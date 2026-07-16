package herdrsock

type WorkspaceCreateParams struct {
	CWD   *string           `json:"cwd,omitempty"`
	Focus bool              `json:"focus,omitempty"`
	Label *string           `json:"label,omitempty"`
	Env   map[string]string `json:"env,omitempty"`
}

type WorkspaceTarget struct {
	WorkspaceID string `json:"workspace_id"`
}

type WorkspaceRenameParams struct {
	WorkspaceID string `json:"workspace_id"`
	Label       string `json:"label"`
}

type WorkspaceMoveParams struct {
	WorkspaceID string `json:"workspace_id"`
	InsertIndex int    `json:"insert_index"`
}

type WorkspaceReportMetadataParams struct {
	WorkspaceID string             `json:"workspace_id"`
	Source      string             `json:"source"`
	Tokens      map[string]*string `json:"tokens"`
	Seq         *uint64            `json:"seq,omitempty"`
	TTLMS       *uint64            `json:"ttl_ms,omitempty"`
}

type WorktreeListParams struct {
	WorkspaceID *string `json:"workspace_id,omitempty"`
	CWD         *string `json:"cwd,omitempty"`
}

type WorktreeCreateParams struct {
	WorkspaceID *string `json:"workspace_id,omitempty"`
	CWD         *string `json:"cwd,omitempty"`
	Branch      *string `json:"branch,omitempty"`
	Base        *string `json:"base,omitempty"`
	Path        *string `json:"path,omitempty"`
	Label       *string `json:"label,omitempty"`
	Focus       bool    `json:"focus,omitempty"`
}

type WorktreeOpenParams struct {
	WorkspaceID *string `json:"workspace_id,omitempty"`
	CWD         *string `json:"cwd,omitempty"`
	Path        *string `json:"path,omitempty"`
	Branch      *string `json:"branch,omitempty"`
	Label       *string `json:"label,omitempty"`
	Focus       bool    `json:"focus,omitempty"`
}

type WorktreeRemoveParams struct {
	WorkspaceID string `json:"workspace_id"`
	Force       bool   `json:"force,omitempty"`
}

type TabCreateParams struct {
	WorkspaceID *string           `json:"workspace_id,omitempty"`
	CWD         *string           `json:"cwd,omitempty"`
	Focus       bool              `json:"focus,omitempty"`
	Label       *string           `json:"label,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
}

type TabListParams struct {
	WorkspaceID *string `json:"workspace_id,omitempty"`
}

type TabTarget struct {
	TabID string `json:"tab_id"`
}

type TabRenameParams struct {
	TabID string `json:"tab_id"`
	Label string `json:"label"`
}

type TabMoveParams struct {
	TabID       string `json:"tab_id"`
	InsertIndex int    `json:"insert_index"`
}

type PaneSplitParams struct {
	WorkspaceID  *string           `json:"workspace_id,omitempty"`
	TargetPaneID *string           `json:"target_pane_id,omitempty"`
	Direction    SplitDirection    `json:"direction"`
	Ratio        *float32          `json:"ratio,omitempty"`
	CWD          *string           `json:"cwd,omitempty"`
	Focus        bool              `json:"focus,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
}

type PaneTarget struct {
	PaneID string `json:"pane_id"`
}

type PaneListParams struct {
	WorkspaceID *string `json:"workspace_id,omitempty"`
}

type PaneCurrentParams struct {
	CallerPaneID *string `json:"caller_pane_id,omitempty"`
}

type PaneRenameParams struct {
	PaneID string  `json:"pane_id"`
	Label  *string `json:"label,omitempty"`
}

type PaneSendTextParams struct {
	PaneID string `json:"pane_id"`
	Text   string `json:"text"`
}

type PaneSendKeysParams struct {
	PaneID string   `json:"pane_id"`
	Keys   []string `json:"keys"`
}

type PaneSendInputParams struct {
	PaneID string   `json:"pane_id"`
	Text   string   `json:"text,omitempty"`
	Keys   []string `json:"keys,omitempty"`
}

type PaneReadParams struct {
	PaneID    string     `json:"pane_id"`
	Source    ReadSource `json:"source"`
	Lines     *uint32    `json:"lines,omitempty"`
	Format    ReadFormat `json:"format,omitempty"`
	StripANSI *bool      `json:"strip_ansi,omitempty"`
}

type PaneReportAgentParams struct {
	PaneID  string         `json:"pane_id"`
	Source  string         `json:"source"`
	Agent   string         `json:"agent"`
	State   PaneAgentState `json:"state"`
	Message *string        `json:"message,omitempty"`
	// Deprecated: Herdr 0.7.4 no longer accepts custom_status. Use State and
	// Message instead.
	CustomStatus     *string `json:"custom_status,omitempty"`
	Seq              *uint64 `json:"seq,omitempty"`
	AgentSessionID   *string `json:"agent_session_id,omitempty"`
	AgentSessionPath *string `json:"agent_session_path,omitempty"`
}

type PaneReportAgentSessionParams struct {
	PaneID             string  `json:"pane_id"`
	Source             string  `json:"source"`
	Agent              string  `json:"agent"`
	Seq                *uint64 `json:"seq,omitempty"`
	AgentSessionID     *string `json:"agent_session_id,omitempty"`
	AgentSessionPath   *string `json:"agent_session_path,omitempty"`
	SessionStartSource *string `json:"session_start_source,omitempty"`
}

type PaneReportMetadataParams struct {
	PaneID          string  `json:"pane_id"`
	Source          string  `json:"source"`
	Agent           *string `json:"agent,omitempty"`
	AppliesToSource *string `json:"applies_to_source,omitempty"`
	Title           *string `json:"title,omitempty"`
	DisplayAgent    *string `json:"display_agent,omitempty"`
	// Deprecated: Herdr 0.7.4 no longer accepts custom_status. Use Title,
	// DisplayAgent, StateLabels, and Tokens instead.
	CustomStatus      *string            `json:"custom_status,omitempty"`
	StateLabels       map[string]string  `json:"state_labels,omitempty"`
	Tokens            map[string]*string `json:"tokens,omitempty"`
	ClearTitle        bool               `json:"clear_title,omitempty"`
	ClearDisplayAgent bool               `json:"clear_display_agent,omitempty"`
	// Deprecated: Herdr 0.7.4 no longer accepts clear_custom_status.
	ClearCustomStatus bool    `json:"clear_custom_status,omitempty"`
	ClearStateLabels  bool    `json:"clear_state_labels,omitempty"`
	Seq               *uint64 `json:"seq,omitempty"`
	TTLMS             *uint64 `json:"ttl_ms,omitempty"`
}

type PaneZoomMode string

const (
	PaneZoomToggle PaneZoomMode = "toggle"
	PaneZoomOn     PaneZoomMode = "on"
	PaneZoomOff    PaneZoomMode = "off"
)

type PaneZoomParams struct {
	PaneID *string      `json:"pane_id,omitempty"`
	Mode   PaneZoomMode `json:"mode,omitempty"`
}

type LayoutSetSplitRatioParams struct {
	TabID  *string `json:"tab_id,omitempty"`
	PaneID *string `json:"pane_id,omitempty"`
	Path   []bool  `json:"path"`
	Ratio  float32 `json:"ratio"`
}

type AgentTarget struct {
	Target string `json:"target"`
}

type AgentReadParams struct {
	Target    string     `json:"target"`
	Source    ReadSource `json:"source"`
	Lines     *uint32    `json:"lines,omitempty"`
	Format    ReadFormat `json:"format,omitempty"`
	StripANSI *bool      `json:"strip_ansi,omitempty"`
}

type AgentSendParams struct {
	Target string `json:"target"`
	Text   string `json:"text"`
}

type AgentStartParams struct {
	Name        string            `json:"name"`
	CWD         *string           `json:"cwd,omitempty"`
	WorkspaceID *string           `json:"workspace_id,omitempty"`
	TabID       *string           `json:"tab_id,omitempty"`
	Split       *SplitDirection   `json:"split,omitempty"`
	Focus       bool              `json:"focus,omitempty"`
	Argv        []string          `json:"argv"`
	Env         map[string]string `json:"env,omitempty"`
}

type NotificationShowParams struct {
	Title    string                `json:"title"`
	Body     *string               `json:"body,omitempty"`
	Position *string               `json:"position,omitempty"`
	Sound    NotificationShowSound `json:"sound,omitempty"`
}
