# P2P Chat

Decentralized chat with end-to-end encryption built on libp2p

## Stack

| Component      | Technology         |
|----------------|--------------------|
| Language       | Go                 |
| Framework      | libp2p             |
| Peer Discovery | Kademlia DHT       |
| Messaging      | Gossipsub (PubSub) |
| Key Derivation | HKDF-SHA256        |
| Encryption     | AES-256-GCM        |

## Features

- **Fully decentralized** — no central server, peers connect directly
- **End-to-end encrypted** — messages encrypted with AES-256-GCM
- **Hidden rooms** — room discovery key, topic and encryption key derived from a single shared secret via HKDF
- **Multiple rooms** — create and join multiple rooms simultaneously
- **Invite system** — share a single invite link to grant room access
- **Three layers of protection**:
    - DHT discovery key hides room existence
    - PubSub topic name hides communication channel
    - AES-GCM encryption hides message content

## Quick Start

### Local

#### Build

```bash
go build -o bin/chat ./cmd/chat
```

#### Run

Terminal 1:

```bash
./bin/chat
```

```
> /create dev
Room 'dev' created
Invite: eyJyb29t...
```

Terminal 2:

```bash
./bin/chat --bootstrap /ip4/127.0.0.1/tcp/PORT/p2p/PEER_ID
```

```
> /join eyJyb29t...
Joined room 'dev' (found 1 peers, connected to 1)
> hello!
```

### Docker

```bash
./deploy/run.sh
```

Starts 3 nodes with automatic bootstrap.

```bash
docker compose -f deploy/docker-compose.yml attach node1

# down
docker compose -f deploy/docker-compose.yml down
```

### Flags

| Flag          | Default      | Description            |
|---------------|--------------|------------------------|
| `--port`      | `0` (random) | Listen port            |
| `--bootstrap` | —            | Bootstrap peer address |

## Commands

| Command          | Description          |
|------------------|----------------------|
| `/create <name>` | Create a new room    |
| `/join <invite>` | Join room via invite |
| `/rooms`         | List joined rooms    |
| `/switch <name>` | Switch active room   |
| `/leave <name>`  | Leave room           |
| `<text>`         | Send message         |

## Architecture

```
cmd/chat/          => entrypoint, CLI
internal/
├── app/           => business logic (create/join/send)
├── config/        => configuration
├── crypto/        => HKDF key derivation, AES-256-GCM
├── domain/        => Room, Message, Invite models
└── p2p/           => libp2p host, DHT, Gossipsub
deploy/            => Dockerfile, docker-compose, run script
```

## Security

```
roomSecret (256-bit random)
       │
       ├── HKDF("DISCOVERY") → discoveryKey  → DHT lookup
       ├── HKDF("TOPIC")     → topicName     → PubSub channel
       └── HKDF("AES")       → aesKey        → message encryption
```

All three keys are derived from a single secret. Without the secret, the room is cryptographically undiscoverable, the
channel is unknown, and messages are unreadable.
