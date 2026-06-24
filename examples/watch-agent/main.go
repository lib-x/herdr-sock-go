package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	herdr "github.com/lib-x/herdr-sock-go"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <pane-id>\n", os.Args[0])
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := herdr.MustNew()
	if _, err := client.RequireCurrentProtocol(ctx); err != nil {
		log.Fatal(err)
	}

	stream, err := client.Subscribe(ctx, herdr.SubscribePaneAgentStatusChanged(os.Args[1], nil))
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for {
		event, err := stream.Next(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Fatal(err)
		}

		var data herdr.PaneAgentStatusChangedEvent
		if err := stream.DecodeData(event, &data); err != nil {
			log.Fatal(err)
		}

		encoded, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
	}
}
