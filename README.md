# herdr-sock-go

Go client for Herdr's local socket API.

This package speaks Herdr's newline-delimited JSON API over the local API
socket. It is generated against Herdr 0.7.0 protocol `14`.

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

## Examples

Run the current pane reader:

```bash
go run ./examples/current-pane
```

Watch agent status changes for a pane:

```bash
go run ./examples/watch-agent w1:p1
```
