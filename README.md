# herdr-sock-go

Go client for Herdr's local socket API.

This package speaks Herdr's newline-delimited JSON API over the local API
socket. It is generated against Herdr 0.7.4 protocol `16`.

Herdr 0.7.3 uses the same protocol and remains compatible for its existing
method set. Metadata token events, JSON pane graphics, and popup panes require
Herdr 0.7.4. The separate binary `pane.graphics.stream` protocol is not part of
this package.

```go
package main

import (
	"context"
	"fmt"
	"log"

	herdr "github.com/lib-x/herdr-sock-go"
)

func main() {
	ctx := context.Background()
	client := herdr.MustNew()

	if _, err := client.RequireCurrentProtocol(ctx); err != nil {
		log.Fatal(err)
	}

	pane, err := client.CurrentPane(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	stripANSI := true
	read, err := client.ReadPane(ctx, herdr.PaneReadParams{
		PaneID:    pane.PaneID,
		Source:    herdr.ReadRecent,
		Format:    herdr.ReadFormatText,
		StripANSI: &stripANSI,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(read.Text)
}
```

## Socket Resolution

`New()` follows Herdr's API socket resolution:

1. `WithSocketPath(path)`
2. `WithSession(name)`
3. `HERDR_SOCKET_PATH`
4. `HERDR_SESSION`
5. the default socket under the Herdr config directory

Use `Call` or `CallRaw` for methods that do not yet have a typed helper:

```go
var result map[string]any
err := client.Call(ctx, "", herdr.MethodPaneProcessInfo, map[string]any{
	"pane_id": "w1:p1",
}, &result)
```

## Metadata

Pane and workspace metadata token patches use `map[string]*string`. A non-nil
pointer sets the value and a nil pointer clears the token:

```go
summary := "reviewing authentication"
ttl := uint64(60_000)
err := client.ReportWorkspaceMetadata(ctx, herdr.WorkspaceReportMetadataParams{
	WorkspaceID: "w1",
	Source:      "my-tool",
	Tokens: map[string]*string{
		"summary": &summary,
		"old":     nil,
	},
	TTLMS: &ttl,
})
```

`WorkspaceInfo`, `PaneInfo`, and `AgentInfo` expose resolved token maps.
`PaneInfo` and `AgentInfo` also expose `TerminalTitle` and
`TerminalTitleStripped`.

## Events

```go
status := herdr.AgentStatusDone
stream, err := client.Subscribe(ctx,
	herdr.SubscribePaneAgentStatusChanged("w1:p1", &status),
)
if err != nil {
	log.Fatal(err)
}
defer stream.Close()

event, err := stream.Next(ctx)
if err != nil {
	log.Fatal(err)
}
fmt.Println(event.Event)
```

Herdr 0.7.4 adds metadata snapshot events:

```go
stream, err := client.Subscribe(ctx,
	herdr.SubscribeWorkspaceMetadataUpdated(),
	herdr.SubscribePaneUpdated(),
)
```

## Popup Plugin Panes

Open a plugin entrypoint as a floating popup with cell or percentage sizing:

```go
placement := herdr.PluginPanePopup
width, err := herdr.PopupPercent(80)
if err != nil {
	log.Fatal(err)
}
height := herdr.PopupCells(24)

pluginPane, err := client.OpenPluginPane(ctx, herdr.PluginPaneOpenParams{
	PluginID:   "pomodoro",
	Entrypoint: "main",
	Placement:  &placement,
	Width:      &width,
	Height:     &height,
	Focus:      true,
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(pluginPane.Pane.PaneID)

defer client.ClosePopup(ctx)
```

Herdr's `ui.copy_on_select` remains a native configuration setting; it is not
a socket API method.

## Examples

Run the current pane reader:

```bash
go run ./examples/current-pane
```

Watch agent status changes for a pane:

```bash
go run ./examples/watch-agent w1:p1
```
