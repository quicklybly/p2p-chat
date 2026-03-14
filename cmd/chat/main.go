package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/quicklybly/p2p-chat/internal/config"
	"github.com/quicklybly/p2p-chat/internal/p2p"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	bootstrap := flag.String("bootstrap", "", "comma-separated bootstrap peer addresses")
	flag.Parse()

	fmt.Println("P2P Chat v0.1")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Load()

	if *bootstrap != "" {
		cfg.P2P.BootstrapPeers = strings.Split(*bootstrap, ",")
	}

	node, err := p2p.NewNode(ctx, cfg.P2P)
	if err != nil {
		log.Fatal(err)
	}
	defer node.Stop()

	node.PrintInfo()

	// graceful shutdown
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("\nRunning. Press Ctrl+C to stop.")
	<-signalChannel
	fmt.Println("\nShutting down...")
}
