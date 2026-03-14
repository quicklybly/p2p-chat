package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"github.com/quicklybly/p2p-chat/internal/config"
	"github.com/quicklybly/p2p-chat/internal/p2p"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	bootstrap := flag.String("bootstrap", "", "comma-separated bootstrap peer addresses")
	room := flag.String("room", "", "room name to provide/find")
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

	testProvideFind(room, node, ctx)

	waitForShutdownSignal()
	fmt.Println("\nShutting down...")
}

func waitForShutdownSignal() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("\nRunning. Press Ctrl+C to stop.")
	<-signalChannel
}

func testProvideFind(room *string, node *p2p.Node, ctx context.Context) {
	if *room != "" {
		go func() {
			key := sha256.Sum256([]byte("room:" + *room))

			// wait for peers before providing
			fmt.Println("\nWaiting for peers in DHT...")
			waitForPeers(ctx, node)

			fmt.Printf("Providing room: %s\n", *room)
			if err := node.Provide(ctx, key[:]); err != nil {
				fmt.Printf("Failed to provide: %s\n", err)
				return
			}
			fmt.Println("Provided successfully")

			time.Sleep(2 * time.Second)

			fmt.Printf("Searching for room: %s\n", *room)
			peers, err := node.FindProviders(ctx, key[:])
			if err != nil {
				fmt.Printf("Failed to find providers: %s\n", err)
				return
			}

			fmt.Printf("Found %d providers:\n", len(peers))
			for _, p := range peers {
				fmt.Printf("  %s\n", p.ID)
			}
		}()
	}
}

func waitForPeers(ctx context.Context, node *p2p.Node) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if len(node.DHT.RoutingTable().ListPeers()) > 0 {
				fmt.Printf("Found %d peers in DHT routing table\n",
					len(node.DHT.RoutingTable().ListPeers()))
				return
			}
			time.Sleep(1 * time.Second)
		}
	}
}
