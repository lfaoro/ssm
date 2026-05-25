// Copyright (c) 2026 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package providers

import (
	"context"
	"fmt"
	"os"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

type GCP struct{}

func (GCP) Name() string { return "gcp" }

func (g GCP) FetchServers(ctx context.Context) ([]Server, error) {
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcp: %w", err)
	}
	defer client.Close()

	projects := gcpProjects()
	if len(projects) == 0 {
		return nil, nil
	}

	var out []Server
	for _, project := range projects {
		req := &computepb.AggregatedListInstancesRequest{
			Project: project,
		}
		it := client.AggregatedList(ctx, req)
		for {
			pair, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("gcp: list instances in %s: %w", project, err)
			}
			for _, inst := range pair.Value.Instances {
				if inst.Status == nil || *inst.Status != "RUNNING" {
					continue
				}
				s := Server{
					Name:     gcpInstanceName(inst),
					Provider: "gcp",
					Region:   gcpRegionFromZone(pair.Key),
					Status:   "running",
				}
				if inst.NetworkInterfaces != nil {
					for _, ni := range inst.NetworkInterfaces {
						if ni.AccessConfigs != nil {
							for _, ac := range ni.AccessConfigs {
								if ac.NatIP != nil {
									s.PublicIP = *ac.NatIP
								}
							}
						}
						if ni.NetworkIP != nil {
							s.PrivateIP = *ni.NetworkIP
						}
					}
				}
				out = append(out, s)
			}
		}
	}
	return out, nil
}

func gcpProjects() []string {
	if p := os.Getenv("GCP_PROJECT"); p != "" {
		return []string{p}
	}
	if p := os.Getenv("GOOGLE_CLOUD_PROJECT"); p != "" {
		return []string{p}
	}
	return nil
}

func gcpInstanceName(inst *computepb.Instance) string {
	if inst.Name != nil {
		return *inst.Name
	}
	return "unknown"
}

func gcpRegionFromZone(zone string) string {
	for i := len(zone) - 1; i >= 0; i-- {
		if zone[i] == '-' {
			return zone[:i]
		}
	}
	return zone
}
