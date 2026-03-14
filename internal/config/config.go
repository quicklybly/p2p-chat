package config

import (
	"fmt"
	"time"
)

type Config struct {
	P2P P2PConfig
}

type P2PConfig struct {
	Port              int
	BootstrapPeers    []string
	EnableMDNS        bool
	ReProvideInterval time.Duration
}

func (config *P2PConfig) ListenAddrs() []string {
	return []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.Port),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", config.Port),
	}
}

func Load() Config {
	return Config{
		P2P: P2PConfig{
			Port:              0,
			BootstrapPeers:    []string{},
			EnableMDNS:        false,
			ReProvideInterval: 10 * time.Minute,
		},
	}
}
