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

	read, err := client.ReadPane(ctx, herdr.PaneReadParams{
		PaneID: pane.PaneID,
		Source: herdr.ReadRecent,
		Format: herdr.ReadFormatText,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("pane: %s\nrevision: %d\n\n%s", pane.PaneID, read.Revision, read.Text)
}
