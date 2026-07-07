package herdrsock

import "context"

type workspaceInfoResult struct {
	Type      string        `json:"type"`
	Workspace WorkspaceInfo `json:"workspace"`
}

type workspaceCreatedResult struct {
	Type      string        `json:"type"`
	Workspace WorkspaceInfo `json:"workspace"`
	Tab       TabInfo       `json:"tab"`
	RootPane  PaneInfo      `json:"root_pane"`
}

type workspaceListResult struct {
	Type       string          `json:"type"`
	Workspaces []WorkspaceInfo `json:"workspaces"`
}

type worktreeListResult struct {
	Type      string             `json:"type"`
	Source    WorktreeSourceInfo `json:"source"`
	Worktrees []WorktreeInfo     `json:"worktrees"`
}

type worktreeCreatedResult struct {
	Type      string        `json:"type"`
	Workspace WorkspaceInfo `json:"workspace"`
	Tab       TabInfo       `json:"tab"`
	RootPane  PaneInfo      `json:"root_pane"`
	Worktree  WorktreeInfo  `json:"worktree"`
}

type worktreeOpenedResult struct {
	Type        string        `json:"type"`
	Workspace   WorkspaceInfo `json:"workspace"`
	Tab         TabInfo       `json:"tab"`
	RootPane    PaneInfo      `json:"root_pane"`
	Worktree    WorktreeInfo  `json:"worktree"`
	AlreadyOpen bool          `json:"already_open"`
}

type WorktreeRemovedResult struct {
	Type        string `json:"type"`
	WorkspaceID string `json:"workspace_id"`
	Path        string `json:"path"`
	Forced      bool   `json:"forced"`
}

type tabInfoResult struct {
	Type string  `json:"type"`
	Tab  TabInfo `json:"tab"`
}

type tabCreatedResult struct {
	Type     string   `json:"type"`
	Tab      TabInfo  `json:"tab"`
	RootPane PaneInfo `json:"root_pane"`
}

type tabListResult struct {
	Type string    `json:"type"`
	Tabs []TabInfo `json:"tabs"`
}

type paneInfoResult struct {
	Type string   `json:"type"`
	Pane PaneInfo `json:"pane"`
}

type paneListResult struct {
	Type  string     `json:"type"`
	Panes []PaneInfo `json:"panes"`
}

type paneCurrentResult struct {
	Type string   `json:"type"`
	Pane PaneInfo `json:"pane"`
}

type paneReadResult struct {
	Type string         `json:"type"`
	Read PaneReadResult `json:"read"`
}

type paneLayoutResult struct {
	Type   string             `json:"type"`
	Layout PaneLayoutSnapshot `json:"layout"`
}

type layoutDescriptionResult struct {
	Type   string            `json:"type"`
	Layout LayoutDescription `json:"layout"`
}

type agentInfoResult struct {
	Type  string    `json:"type"`
	Agent AgentInfo `json:"agent"`
}

type agentListResult struct {
	Type   string      `json:"type"`
	Agents []AgentInfo `json:"agents"`
}

type agentStartedResult struct {
	Type  string    `json:"type"`
	Agent AgentInfo `json:"agent"`
	Argv  []string  `json:"argv"`
}

type NotificationShowResult struct {
	Type   string `json:"type"`
	Shown  bool   `json:"shown"`
	Reason string `json:"reason"`
}

type sessionSnapshotResult struct {
	Type     string          `json:"type"`
	Snapshot SessionSnapshot `json:"snapshot"`
}

func (c *Client) SessionSnapshot(ctx context.Context) (*SessionSnapshot, error) {
	var out sessionSnapshotResult
	if err := c.Call(ctx, "", MethodSessionSnapshot, EmptyParams{}, &out); err != nil {
		return nil, err
	}
	return &out.Snapshot, nil
}

// CreateWorkspace creates a workspace and returns the workspace, first tab,
// and root pane records.
func (c *Client) CreateWorkspace(ctx context.Context, params WorkspaceCreateParams) (*WorkspaceInfo, *TabInfo, *PaneInfo, error) {
	var out workspaceCreatedResult
	if err := c.Call(ctx, "", MethodWorkspaceCreate, params, &out); err != nil {
		return nil, nil, nil, err
	}
	return &out.Workspace, &out.Tab, &out.RootPane, nil
}

func (c *Client) ListWorkspaces(ctx context.Context) ([]WorkspaceInfo, error) {
	var out workspaceListResult
	if err := c.Call(ctx, "", MethodWorkspaceList, EmptyParams{}, &out); err != nil {
		return nil, err
	}
	return out.Workspaces, nil
}

func (c *Client) GetWorkspace(ctx context.Context, workspaceID string) (*WorkspaceInfo, error) {
	var out workspaceInfoResult
	if err := c.Call(ctx, "", MethodWorkspaceGet, WorkspaceTarget{WorkspaceID: workspaceID}, &out); err != nil {
		return nil, err
	}
	return &out.Workspace, nil
}

func (c *Client) FocusWorkspace(ctx context.Context, workspaceID string) (*WorkspaceInfo, error) {
	var out workspaceInfoResult
	if err := c.Call(ctx, "", MethodWorkspaceFocus, WorkspaceTarget{WorkspaceID: workspaceID}, &out); err != nil {
		return nil, err
	}
	return &out.Workspace, nil
}

func (c *Client) RenameWorkspace(ctx context.Context, workspaceID, label string) (*WorkspaceInfo, error) {
	var out workspaceInfoResult
	params := WorkspaceRenameParams{WorkspaceID: workspaceID, Label: label}
	if err := c.Call(ctx, "", MethodWorkspaceRename, params, &out); err != nil {
		return nil, err
	}
	return &out.Workspace, nil
}

func (c *Client) MoveWorkspace(ctx context.Context, workspaceID string, insertIndex int) ([]WorkspaceInfo, error) {
	var out workspaceListResult
	params := WorkspaceMoveParams{WorkspaceID: workspaceID, InsertIndex: insertIndex}
	if err := c.Call(ctx, "", MethodWorkspaceMove, params, &out); err != nil {
		return nil, err
	}
	return out.Workspaces, nil
}

func (c *Client) CloseWorkspace(ctx context.Context, workspaceID string) error {
	return c.Call(ctx, "", MethodWorkspaceClose, WorkspaceTarget{WorkspaceID: workspaceID}, nil)
}

func (c *Client) ListWorktrees(ctx context.Context, params WorktreeListParams) (*WorktreeSourceInfo, []WorktreeInfo, error) {
	var out worktreeListResult
	if err := c.Call(ctx, "", MethodWorktreeList, params, &out); err != nil {
		return nil, nil, err
	}
	return &out.Source, out.Worktrees, nil
}

func (c *Client) CreateWorktree(ctx context.Context, params WorktreeCreateParams) (*WorkspaceInfo, *TabInfo, *PaneInfo, *WorktreeInfo, error) {
	var out worktreeCreatedResult
	if err := c.Call(ctx, "", MethodWorktreeCreate, params, &out); err != nil {
		return nil, nil, nil, nil, err
	}
	return &out.Workspace, &out.Tab, &out.RootPane, &out.Worktree, nil
}

func (c *Client) OpenWorktree(ctx context.Context, params WorktreeOpenParams) (*WorkspaceInfo, *TabInfo, *PaneInfo, *WorktreeInfo, bool, error) {
	var out worktreeOpenedResult
	if err := c.Call(ctx, "", MethodWorktreeOpen, params, &out); err != nil {
		return nil, nil, nil, nil, false, err
	}
	return &out.Workspace, &out.Tab, &out.RootPane, &out.Worktree, out.AlreadyOpen, nil
}

func (c *Client) RemoveWorktree(ctx context.Context, params WorktreeRemoveParams) (*WorktreeRemovedResult, error) {
	var out WorktreeRemovedResult
	if err := c.Call(ctx, "", MethodWorktreeRemove, params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateTab(ctx context.Context, params TabCreateParams) (*TabInfo, *PaneInfo, error) {
	var out tabCreatedResult
	if err := c.Call(ctx, "", MethodTabCreate, params, &out); err != nil {
		return nil, nil, err
	}
	return &out.Tab, &out.RootPane, nil
}

func (c *Client) ListTabs(ctx context.Context, workspaceID *string) ([]TabInfo, error) {
	var out tabListResult
	if err := c.Call(ctx, "", MethodTabList, TabListParams{WorkspaceID: workspaceID}, &out); err != nil {
		return nil, err
	}
	return out.Tabs, nil
}

func (c *Client) GetTab(ctx context.Context, tabID string) (*TabInfo, error) {
	var out tabInfoResult
	if err := c.Call(ctx, "", MethodTabGet, TabTarget{TabID: tabID}, &out); err != nil {
		return nil, err
	}
	return &out.Tab, nil
}

func (c *Client) FocusTab(ctx context.Context, tabID string) (*TabInfo, error) {
	var out tabInfoResult
	if err := c.Call(ctx, "", MethodTabFocus, TabTarget{TabID: tabID}, &out); err != nil {
		return nil, err
	}
	return &out.Tab, nil
}

func (c *Client) RenameTab(ctx context.Context, tabID, label string) (*TabInfo, error) {
	var out tabInfoResult
	if err := c.Call(ctx, "", MethodTabRename, TabRenameParams{TabID: tabID, Label: label}, &out); err != nil {
		return nil, err
	}
	return &out.Tab, nil
}

func (c *Client) MoveTab(ctx context.Context, tabID string, insertIndex int) ([]TabInfo, error) {
	var out tabListResult
	params := TabMoveParams{TabID: tabID, InsertIndex: insertIndex}
	if err := c.Call(ctx, "", MethodTabMove, params, &out); err != nil {
		return nil, err
	}
	return out.Tabs, nil
}

func (c *Client) CloseTab(ctx context.Context, tabID string) error {
	return c.Call(ctx, "", MethodTabClose, TabTarget{TabID: tabID}, nil)
}

func (c *Client) CurrentPane(ctx context.Context, callerPaneID *string) (*PaneInfo, error) {
	var out paneCurrentResult
	if err := c.Call(ctx, "", MethodPaneCurrent, PaneCurrentParams{CallerPaneID: callerPaneID}, &out); err != nil {
		return nil, err
	}
	return &out.Pane, nil
}

func (c *Client) GetPane(ctx context.Context, paneID string) (*PaneInfo, error) {
	var out paneInfoResult
	if err := c.Call(ctx, "", MethodPaneGet, PaneTarget{PaneID: paneID}, &out); err != nil {
		return nil, err
	}
	return &out.Pane, nil
}

func (c *Client) FocusPane(ctx context.Context, paneID string) (*PaneInfo, error) {
	var out paneInfoResult
	if err := c.Call(ctx, "", MethodPaneFocus, PaneTarget{PaneID: paneID}, &out); err != nil {
		return nil, err
	}
	return &out.Pane, nil
}

func (c *Client) ListPanes(ctx context.Context, workspaceID *string) ([]PaneInfo, error) {
	var out paneListResult
	if err := c.Call(ctx, "", MethodPaneList, PaneListParams{WorkspaceID: workspaceID}, &out); err != nil {
		return nil, err
	}
	return out.Panes, nil
}

func (c *Client) SplitPane(ctx context.Context, params PaneSplitParams) (*PaneInfo, error) {
	var out paneInfoResult
	if err := c.Call(ctx, "", MethodPaneSplit, params, &out); err != nil {
		return nil, err
	}
	return &out.Pane, nil
}

func (c *Client) RenamePane(ctx context.Context, paneID string, label *string) (*PaneInfo, error) {
	var out paneInfoResult
	if err := c.Call(ctx, "", MethodPaneRename, PaneRenameParams{PaneID: paneID, Label: label}, &out); err != nil {
		return nil, err
	}
	return &out.Pane, nil
}

func (c *Client) PaneLayout(ctx context.Context, paneID *string) (*PaneLayoutSnapshot, error) {
	var out paneLayoutResult
	params := struct {
		PaneID *string `json:"pane_id,omitempty"`
	}{PaneID: paneID}
	if err := c.Call(ctx, "", MethodPaneLayout, params, &out); err != nil {
		return nil, err
	}
	return &out.Layout, nil
}

func (c *Client) SetSplitRatio(ctx context.Context, params LayoutSetSplitRatioParams) (*LayoutDescription, error) {
	var out layoutDescriptionResult
	if err := c.Call(ctx, "", MethodLayoutSetSplitRatio, params, &out); err != nil {
		return nil, err
	}
	return &out.Layout, nil
}

func (c *Client) ZoomPane(ctx context.Context, params PaneZoomParams) error {
	return c.Call(ctx, "", MethodPaneZoom, params, nil)
}

func (c *Client) ClosePane(ctx context.Context, paneID string) error {
	return c.Call(ctx, "", MethodPaneClose, PaneTarget{PaneID: paneID}, nil)
}

func (c *Client) SendText(ctx context.Context, paneID, text string) error {
	return c.Call(ctx, "", MethodPaneSendText, PaneSendTextParams{PaneID: paneID, Text: text}, nil)
}

func (c *Client) SendKeys(ctx context.Context, paneID string, keys ...string) error {
	return c.Call(ctx, "", MethodPaneSendKeys, PaneSendKeysParams{PaneID: paneID, Keys: keys}, nil)
}

func (c *Client) SendInput(ctx context.Context, params PaneSendInputParams) error {
	return c.Call(ctx, "", MethodPaneSendInput, params, nil)
}

func (c *Client) ReadPane(ctx context.Context, params PaneReadParams) (*PaneReadResult, error) {
	if params.Format == "" {
		params.Format = ReadFormatText
	}
	var out paneReadResult
	if err := c.Call(ctx, "", MethodPaneRead, params, &out); err != nil {
		return nil, err
	}
	return &out.Read, nil
}

func (c *Client) ReportAgent(ctx context.Context, params PaneReportAgentParams) error {
	return c.Call(ctx, "", MethodPaneReportAgent, params, nil)
}

func (c *Client) ReportAgentSession(ctx context.Context, params PaneReportAgentSessionParams) error {
	return c.Call(ctx, "", MethodPaneReportAgentSession, params, nil)
}

func (c *Client) ReportMetadata(ctx context.Context, params PaneReportMetadataParams) error {
	return c.Call(ctx, "", MethodPaneReportMetadata, params, nil)
}

func (c *Client) ListAgents(ctx context.Context) ([]AgentInfo, error) {
	var out agentListResult
	if err := c.Call(ctx, "", MethodAgentList, EmptyParams{}, &out); err != nil {
		return nil, err
	}
	return out.Agents, nil
}

func (c *Client) GetAgent(ctx context.Context, target string) (*AgentInfo, error) {
	var out agentInfoResult
	if err := c.Call(ctx, "", MethodAgentGet, AgentTarget{Target: target}, &out); err != nil {
		return nil, err
	}
	return &out.Agent, nil
}

func (c *Client) SendAgent(ctx context.Context, target, text string) error {
	return c.Call(ctx, "", MethodAgentSend, AgentSendParams{Target: target, Text: text}, nil)
}

func (c *Client) StartAgent(ctx context.Context, params AgentStartParams) (*AgentInfo, []string, error) {
	var out agentStartedResult
	if err := c.Call(ctx, "", MethodAgentStart, params, &out); err != nil {
		return nil, nil, err
	}
	return &out.Agent, out.Argv, nil
}

func (c *Client) ShowNotification(ctx context.Context, params NotificationShowParams) (*NotificationShowResult, error) {
	var out NotificationShowResult
	if err := c.Call(ctx, "", MethodNotificationShow, params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
