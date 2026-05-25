// Copyright (c) 2026 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

// Package providers defines interfaces for fetching servers from cloud providers.
package providers

import "context"

// Server represents a single cloud server/instance discovered from a provider.
type Server struct {
	Name      string
	PublicIP  string
	PrivateIP string
	Provider  string
	Region    string
	Status    string
}

// Provider fetches servers from a single cloud provider.
type Provider interface {
	// Name returns the provider identifier ("hetzner", "aws", "gcp", "azure").
	Name() string
	// FetchServers returns all running servers accessible by the configured credentials.
	FetchServers(ctx context.Context) ([]Server, error)
}
