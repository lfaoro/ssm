// Copyright (c) 2026 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package providers

import (
	"context"
	"os"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type Hetzner struct{}

func (Hetzner) Name() string { return "hetzner" }

func (h Hetzner) FetchServers(ctx context.Context) ([]Server, error) {
	token := os.Getenv("HCLOUD_TOKEN")
	if token == "" {
		return nil, nil
	}
	client := hcloud.NewClient(hcloud.WithToken(token))
	servers, err := client.Server.All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Server, 0, len(servers))
	for _, s := range servers {
		if s.Status != hcloud.ServerStatusRunning {
			continue
		}
		var pubIP, privIP string
		if s.PublicNet.IPv4.IP != nil {
			pubIP = s.PublicNet.IPv4.IP.String()
		}
		if s.PrivateNet != nil && len(s.PrivateNet) > 0 {
			if s.PrivateNet[0].IP != nil {
				privIP = s.PrivateNet[0].IP.String()
			}
		}
		out = append(out, Server{
			Name:      s.Name,
			PublicIP:  pubIP,
			PrivateIP: privIP,
			Provider:  "hetzner",
			Region:    s.Datacenter.Location.Name,
			Status:    string(s.Status),
		})
	}
	return out, nil
}
