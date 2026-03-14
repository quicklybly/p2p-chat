package main

import (
	"bufio"
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
	port := flag.Int("port", 0, "listen port (0 for random)")
	flag.Parse()

	fmt.Println("P2P Chat v0.1")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Load()
	cfg.P2P.Port = *port

	if *bootstrap != "" {
		cfg.P2P.BootstrapPeers = strings.Split(*bootstrap, ",")
	}

	node, err := p2p.NewNode(ctx, cfg.P2P)
	if err != nil {
		log.Fatal(err)
	}
	defer node.Stop()

	node.PrintInfo()

	if *room == "" {
		log.Fatal("--room is required")
	}

	testProvideFind(room, node, ctx)
	launchPubSub(ctx, node, room)

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
	go func() {
		key := sha256.Sum256([]byte("room:" + *room))

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
			fmt.Printf("  Connecting to %s\n", p.ID)
			if err := node.ConnectToPeer(ctx, p); err != nil {
				fmt.Printf("  Failed: %s\n", err)
			} else {
				fmt.Printf("  Connected to %s\n", p.ID)
			}
		}
	}()
}

func launchPubSub(ctx context.Context, node *p2p.Node, room *string) {
	topic, err := node.PubSub.Join(*room)
	if err != nil {
		log.Fatal(err)
	}

	sub, err := node.PubSub.Subscribe(topic)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nJoined room: %s\n", *room)
	fmt.Println("Type a message and press Enter to send.\n")

	// read incoming messages
	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				return
			}
			if msg.ReceivedFrom == node.ID() {
				continue
			}
			fmt.Printf("\n[%s]: %s\n> ", msg.ReceivedFrom.String()[:8], string(msg.Data))
		}
	}()

	// read stdin and publish
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")

	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			if text == "" {
				fmt.Print("> ")
				continue
			}
			if err := node.PubSub.Publish(ctx, topic, []byte(text)); err != nil {
				fmt.Printf("Failed to send: %s\n", err)
			}
			fmt.Print("> ")
		}
	}()
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
