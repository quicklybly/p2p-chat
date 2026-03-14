package config

import "time"

type Config struct {
	P2P P2PConfig
}

type P2PConfig struct {
	ListenAddrs       []string
	BootstrapPeers    []string
	EnableMDNS        bool
	ReProvideInterval time.Duration
}

func Load() Config {
	return Config{
		P2P: P2PConfig{
			ListenAddrs: []string{
				"/ip4/0.0.0.0/tcp/0",
				"/ip4/0.0.0.0/udp/0/quic-v1",
			},
			BootstrapPeers:    []string{},
			EnableMDNS:        false,
			ReProvideInterval: 10 * time.Minute,
		},
	}
}
