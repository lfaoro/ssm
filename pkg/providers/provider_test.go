// Copyright (c) 2026 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package providers

import "testing"

var _ Provider = Hetzner{}
var _ Provider = AWS{}
var _ Provider = GCP{}
var _ Provider = Azure{}

func TestAllProvidersImplementInterface(t *testing.T) {
	providers := []Provider{Hetzner{}, AWS{}, GCP{}, Azure{}}
	for _, p := range providers {
		if p.Name() == "" {
			t.Errorf("%T has empty Name", p)
		}
	}
}
