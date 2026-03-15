#!/bin/bash
set -e

cd "$(dirname "$0")/.."

echo "=== Building binary ==="
GOOS=linux GOARCH=arm64 go build -o deploy/chat ./cmd/chat

echo "=== Building Docker image ==="
docker compose -f deploy/docker-compose.yml build

echo "=== Stopping old containers ==="
docker compose -f deploy/docker-compose.yml down 2>/dev/null || true

echo "=== Starting node1 ==="
docker compose -f deploy/docker-compose.yml up -d node1
sleep 3

echo "=== Getting node1 PeerID ==="
NODE1_ID=$(docker compose -f deploy/docker-compose.yml logs --no-color node1 2>&1 | grep "PeerID:" | head -1 | awk '{print $NF}' | tr -d '\r\n ')

if [ -z "$NODE1_ID" ]; then
    echo "ERROR: Could not get node1 PeerID"
    docker compose -f deploy/docker-compose.yml logs node1
    exit 1
fi

echo "Node1 PeerID: $NODE1_ID"

BOOTSTRAP="/dns4/node1/tcp/9000/p2p/$NODE1_ID"
echo "Bootstrap: $BOOTSTRAP"

echo "=== Starting node2 and node3 ==="
BOOTSTRAP=$BOOTSTRAP docker compose -f deploy/docker-compose.yml up -d node2 node3
sleep 3

echo ""
echo "=== All nodes running ==="
echo ""
echo "Usage:"
echo "  1. Attach to node1:  docker compose -f deploy/docker-compose.yml attach node1"
echo "  2. Create room:      /create dev"
echo "  3. Copy invite"
echo "  4. Attach to node2:  docker compose -f deploy/docker-compose.yml attach node2"
echo "  5. Join room:        /join <invite>"
echo "  6. Send messages"
echo ""
echo "  Detach: Ctrl+P, Ctrl+Q"
echo "  Logs:   docker compose -f deploy/docker-compose.yml logs -f"
echo "  Stop:   docker compose -f deploy/docker-compose.yml down"
