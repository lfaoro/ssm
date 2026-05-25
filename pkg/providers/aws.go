// Copyright (c) 2026 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package providers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// AWS fetches EC2 instances across all regions using the standard SDK credential chain.
type AWS struct{}

func (AWS) Name() string { return "aws" } //nolint:revive

// FetchServers returns all running EC2 instances across all regions.
func (a AWS) FetchServers(ctx context.Context) ([]Server, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("aws: %w", err)
	}
	regions, err := ec2.NewFromConfig(cfg).DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("aws: listing regions: %w", err)
	}
	var out []Server
	for _, region := range regions.Regions {
		cfg.Region = *region.RegionName
		svc := ec2.NewFromConfig(cfg)
		var nextToken *string
		for {
			resp, err := svc.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
				NextToken: nextToken,
				Filters: []types.Filter{{
					Name:   aws.String("instance-state-name"),
					Values: []string{"running"},
				}},
			})
			if err != nil {
				return nil, fmt.Errorf("aws: describe instances in %s: %w", *region.RegionName, err)
			}
			for _, r := range resp.Reservations {
				for _, inst := range r.Instances {
					s := Server{
						Name:     *inst.InstanceId,
						Provider: "aws",
						Region:   *region.RegionName,
						Status:   string(inst.State.Name),
					}
					if inst.Tags != nil {
						for _, t := range inst.Tags {
							if t.Key != nil && *t.Key == "Name" && t.Value != nil {
								s.Name = *t.Value
							}
						}
					}
					if inst.PublicIpAddress != nil {
						s.PublicIP = *inst.PublicIpAddress
					}
					if inst.PrivateIpAddress != nil {
						s.PrivateIP = *inst.PrivateIpAddress
					}
					out = append(out, s)
				}
			}
			nextToken = resp.NextToken
			if nextToken == nil || *nextToken == "" {
				break
			}
		}
	}
	return out, nil
}
