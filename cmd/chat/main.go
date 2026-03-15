package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/quicklybly/p2p-chat/internal/app"
	"github.com/quicklybly/p2p-chat/internal/config"
	"github.com/quicklybly/p2p-chat/internal/domain"
	"github.com/quicklybly/p2p-chat/internal/p2p"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	bootstrap := flag.String("bootstrap", "", "bootstrap peer address")
	port := flag.Int("port", 0, "listen port (0 = random)")
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

	if len(cfg.P2P.BootstrapPeers) > 0 {
		waitForPeers(ctx, node)
	}

	service := app.NewService(node)

	service.OnMessage(func(msg domain.Message) {
		fmt.Printf("\n[%s] %s: %s\n> ",
			msg.RoomName,
			msg.SenderID[:8],
			msg.Text)
	})

	printHelp()

	var activeRoom string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")

	cli(scanner, activeRoom, service, ctx)

	waitForShutdownSignal()
	fmt.Println("\nShutting down...")
}

func waitForShutdownSignal() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("\nRunning. Press Ctrl+C to stop.")
	<-signalChannel
}

func printHelp() {
	fmt.Println("\nCommands:")
	fmt.Println("  /create <name>  - create room")
	fmt.Println("  /join <invite>  - join room")
	fmt.Println("  /rooms          - list rooms")
	fmt.Println("  /switch <name>  - switch room")
	fmt.Println("  /leave <name>   - leave room")
	fmt.Println("  <text>          - send message")
	fmt.Println()
}

func cli(scanner *bufio.Scanner, activeRoom string, service *app.Service, ctx context.Context) {
	go func() {
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				fmt.Print("> ")
				continue
			}
			switch {

			case strings.HasPrefix(line, "/create "):
				name := strings.TrimPrefix(line, "/create ")
				invite, err := service.CreateRoom(ctx, name)
				if err != nil {
					fmt.Printf("Error: %s\n", err)
				} else {
					activeRoom = name
					fmt.Printf("Room '%s' created\n", name)
					fmt.Printf("Invite: %s\n", invite)
				}

			case strings.HasPrefix(line, "/join "):
				inviteStr := strings.TrimPrefix(line, "/join ")
				roomName, err := service.JoinRoom(ctx, inviteStr)
				if err != nil {
					fmt.Printf("Error: %s\n", err)
				} else {
					activeRoom = roomName
				}

			case line == "/rooms":
				rooms := service.Rooms()
				if len(rooms) == 0 {
					fmt.Println("No rooms")
				}
				for _, r := range rooms {
					marker := " "
					if r == activeRoom {
						marker = "*"
					}
					fmt.Printf("  %s %s\n", marker, r)
				}

			case strings.HasPrefix(line, "/switch "):
				name := strings.TrimPrefix(line, "/switch ")
				if service.HasRoom(name) {
					activeRoom = name
					fmt.Printf("Switched to '%s'\n", name)
				} else {
					fmt.Printf("Room '%s' not found\n", name)
				}

			case strings.HasPrefix(line, "/leave "):
				name := strings.TrimPrefix(line, "/leave ")
				if err := service.LeaveRoom(name); err != nil {
					fmt.Printf("Error: %s\n", err)
				} else {
					fmt.Printf("Left room '%s'\n", name)
					if activeRoom == name {
						activeRoom = ""
					}
				}

			default:
				if activeRoom == "" {
					fmt.Println("No active room. Use /create or /join")
				} else {
					if err := service.SendMessage(ctx, activeRoom, line); err != nil {
						fmt.Printf("Error: %s\n", err)
					}
				}
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
