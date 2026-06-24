package herdrsock

type AgentStatus string

const (
	AgentStatusIdle    AgentStatus = "idle"
	AgentStatusWorking AgentStatus = "working"
	AgentStatusBlocked AgentStatus = "blocked"
	AgentStatusDone    AgentStatus = "done"
	AgentStatusUnknown AgentStatus = "unknown"
)

type PaneAgentState string

const (
	PaneAgentStateIdle    PaneAgentState = "idle"
	PaneAgentStateWorking PaneAgentState = "working"
	PaneAgentStateBlocked PaneAgentState = "blocked"
	PaneAgentStateUnknown PaneAgentState = "unknown"
)

type SplitDirection string

const (
	SplitRight SplitDirection = "right"
	SplitDown  SplitDirection = "down"
)

type PaneDirection string

const (
	PaneLeft  PaneDirection = "left"
	PaneRight PaneDirection = "right"
	PaneUp    PaneDirection = "up"
	PaneDown  PaneDirection = "down"
)

type ReadSource string

const (
	ReadVisible         ReadSource = "visible"
	ReadRecent          ReadSource = "recent"
	ReadRecentUnwrapped ReadSource = "recent_unwrapped"
	ReadDetection       ReadSource = "detection"
)

type ReadFormat string

const (
	ReadFormatText ReadFormat = "text"
	ReadFormatANSI ReadFormat = "ansi"
)

type PongResult struct {
	Type         string              `json:"type"`
	Version      string              `json:"version"`
	Protocol     uint32              `json:"protocol"`
	Capabilities *ServerCapabilities `json:"capabilities,omitempty"`
}

type ServerCapabilities struct {
	LiveHandoff bool `json:"live_handoff"`
}

type ConfigReloadResult struct {
	Type        string   `json:"type"`
	Status      string   `json:"status"`
	Diagnostics []string `json:"diagnostics"`
}

type WorkspaceInfo struct {
	WorkspaceID string                 `json:"workspace_id"`
	Number      int                    `json:"number"`
	Label       string                 `json:"label"`
	Focused     bool                   `json:"focused"`
	PaneCount   int                    `json:"pane_count"`
	TabCount    int                    `json:"tab_count"`
	ActiveTabID string                 `json:"active_tab_id"`
	AgentStatus AgentStatus            `json:"agent_status"`
	Worktree    *WorkspaceWorktreeInfo `json:"worktree,omitempty"`
}

type WorkspaceWorktreeInfo struct {
	RepoKey          string `json:"repo_key"`
	RepoName         string `json:"repo_name"`
	RepoRoot         string `json:"repo_root"`
	CheckoutPath     string `json:"checkout_path"`
	IsLinkedWorktree bool   `json:"is_linked_worktree"`
}

type TabInfo struct {
	TabID       string      `json:"tab_id"`
	WorkspaceID string      `json:"workspace_id"`
	Number      int         `json:"number"`
	Label       string      `json:"label"`
	Focused     bool        `json:"focused"`
	PaneCount   int         `json:"pane_count"`
	AgentStatus AgentStatus `json:"agent_status"`
}

type AgentSessionInfo struct {
	Source string `json:"source"`
	Agent  string `json:"agent"`
	Kind   string `json:"kind"`
	Value  string `json:"value"`
}

type PaneInfo struct {
	PaneID        string            `json:"pane_id"`
	TerminalID    string            `json:"terminal_id"`
	WorkspaceID   string            `json:"workspace_id"`
	TabID         string            `json:"tab_id"`
	Focused       bool              `json:"focused"`
	CWD           *string           `json:"cwd,omitempty"`
	ForegroundCWD *string           `json:"foreground_cwd,omitempty"`
	Label         *string           `json:"label,omitempty"`
	Agent         *string           `json:"agent,omitempty"`
	Title         *string           `json:"title,omitempty"`
	DisplayAgent  *string           `json:"display_agent,omitempty"`
	AgentStatus   AgentStatus       `json:"agent_status"`
	CustomStatus  *string           `json:"custom_status,omitempty"`
	StateLabels   map[string]string `json:"state_labels,omitempty"`
	AgentSession  *AgentSessionInfo `json:"agent_session,omitempty"`
	Revision      uint64            `json:"revision"`
}

type AgentInfo struct {
	TerminalID             string            `json:"terminal_id"`
	Name                   *string           `json:"name,omitempty"`
	Agent                  *string           `json:"agent,omitempty"`
	Title                  *string           `json:"title,omitempty"`
	DisplayAgent           *string           `json:"display_agent,omitempty"`
	AgentStatus            AgentStatus       `json:"agent_status"`
	ScreenDetectionSkipped bool              `json:"screen_detection_skipped,omitempty"`
	CustomStatus           *string           `json:"custom_status,omitempty"`
	StateLabels            map[string]string `json:"state_labels,omitempty"`
	AgentSession           *AgentSessionInfo `json:"agent_session,omitempty"`
	WorkspaceID            string            `json:"workspace_id"`
	TabID                  string            `json:"tab_id"`
	PaneID                 string            `json:"pane_id"`
	Focused                bool              `json:"focused"`
	CWD                    *string           `json:"cwd,omitempty"`
	ForegroundCWD          *string           `json:"foreground_cwd,omitempty"`
	Revision               uint64            `json:"revision"`
}

type PaneReadResult struct {
	PaneID      string     `json:"pane_id"`
	WorkspaceID string     `json:"workspace_id"`
	TabID       string     `json:"tab_id"`
	Source      ReadSource `json:"source"`
	Format      ReadFormat `json:"format"`
	Text        string     `json:"text"`
	Revision    uint64     `json:"revision"`
	Truncated   bool       `json:"truncated"`
}

type PaneLayoutRect struct {
	X      uint16 `json:"x"`
	Y      uint16 `json:"y"`
	Width  uint16 `json:"width"`
	Height uint16 `json:"height"`
}

type PaneLayoutSnapshot struct {
	WorkspaceID   string            `json:"workspace_id"`
	TabID         string            `json:"tab_id"`
	Zoomed        bool              `json:"zoomed"`
	Area          PaneLayoutRect    `json:"area"`
	FocusedPaneID string            `json:"focused_pane_id"`
	Panes         []PaneLayoutPane  `json:"panes"`
	Splits        []PaneLayoutSplit `json:"splits"`
}

type PaneLayoutPane struct {
	PaneID  string         `json:"pane_id"`
	Focused bool           `json:"focused"`
	Rect    PaneLayoutRect `json:"rect"`
}

type PaneLayoutSplit struct {
	ID        string         `json:"id"`
	Direction SplitDirection `json:"direction"`
	Ratio     float32        `json:"ratio"`
	Rect      PaneLayoutRect `json:"rect"`
}

type WorktreeSourceInfo struct {
	RepoKey            string  `json:"repo_key"`
	RepoName           string  `json:"repo_name"`
	RepoRoot           string  `json:"repo_root"`
	SourceCheckoutPath string  `json:"source_checkout_path"`
	SourceWorkspaceID  *string `json:"source_workspace_id,omitempty"`
}

type WorktreeInfo struct {
	Path             string  `json:"path"`
	Branch           *string `json:"branch,omitempty"`
	IsBare           bool    `json:"is_bare"`
	IsDetached       bool    `json:"is_detached"`
	IsPrunable       bool    `json:"is_prunable"`
	IsLinkedWorktree bool    `json:"is_linked_worktree"`
	OpenWorkspaceID  *string `json:"open_workspace_id,omitempty"`
	Label            string  `json:"label"`
}

type NotificationShowSound string

const (
	NotificationSoundNone    NotificationShowSound = "none"
	NotificationSoundDone    NotificationShowSound = "done"
	NotificationSoundRequest NotificationShowSound = "request"
)
