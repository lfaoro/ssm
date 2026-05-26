// Copyright (c) 2026 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package providers

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// Azure fetches virtual machines from the subscription identified by AZURE_SUBSCRIPTION_ID.
type Azure struct{}

func (Azure) Name() string { return "azure" } //nolint:revive

// FetchServers returns all running Azure VMs in the configured subscription.
func (a Azure) FetchServers(ctx context.Context) ([]Server, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("azure: credentials: %w", err)
	}
	subID := azureSubscription()
	if subID == "" {
		return nil, nil
	}
	rgClient, err := armresources.NewResourceGroupsClient(subID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("azure: resource groups: %w", err)
	}
	vmClient, err := armcompute.NewVirtualMachinesClient(subID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("azure: vm client: %w", err)
	}
	var out []Server
	pager := rgClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("azure: list resource groups: %w", err)
		}
		for _, rg := range page.Value {
			if rg.Name == nil {
				continue
			}
			vmPager := vmClient.NewListPager(*rg.Name, nil)
			for vmPager.More() {
				vmPage, err := vmPager.NextPage(ctx)
				if err != nil {
					return nil, fmt.Errorf("azure: list vms in %s: %w", *rg.Name, err)
				}
				for _, vm := range vmPage.Value {
					if vm.Name == nil {
						continue
					}
					s := Server{
						Name:     *vm.Name,
						Provider: "azure",
						Region:   azureRegion(vm),
						Status:   "running",
					}
					out = append(out, s)
				}
			}
		}
	}
	return out, nil
}

func azureSubscription() string {
	if v := os.Getenv("AZURE_SUBSCRIPTION_ID"); v != "" {
		return v
	}
	return os.Getenv("ARM_SUBSCRIPTION_ID")
}

func azureRegion(vm *armcompute.VirtualMachine) string {
	if vm.Location != nil {
		return *vm.Location
	}
	return ""
}
