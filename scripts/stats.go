// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

// Fetch download counts from GitHub releases API and write data/stats.json.
//
// Usage: go run scripts/stats.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	repo    = "lfaoro/ssm"
	apiURL  = "https://api.github.com/repos/" + repo + "/releases?per_page=100"
)

func main() {
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stats: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ssm-stats/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stats: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stats: %v\n", err)
		os.Exit(1)
	}

	var releases []struct {
		Assets []struct {
			Name         string `json:"name"`
			DownloadCount int   `json:"download_count"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(body, &releases); err != nil {
		fmt.Fprintf(os.Stderr, "stats: %v\n", err)
		os.Exit(1)
	}

	knownOS := []string{"freebsd", "netbsd", "openbsd", "solaris", "windows", "linux", "darwin"}
	platforms := make(map[string]int)
	var total int

	for _, rel := range releases {
		for _, asset := range rel.Assets {
			name := asset.Name
			if strings.HasSuffix(name, ".asc") || strings.HasSuffix(name, "checksums.txt") {
				continue
			}
			stem := name
			for _, ext := range []string{".tar.gz", ".tgz", ".zip", ".deb", ".rpm"} {
				if strings.HasSuffix(stem, ext) {
					stem = stem[:len(stem)-len(ext)]
					break
				}
			}

			key := stem
			for _, osName := range knownOS {
				idx := strings.Index(stem, osName+"_")
				if idx >= 0 {
					key = osName + "/" + stem[idx+len(osName)+1:]
					break
				}
			}
			platforms[key] += asset.DownloadCount
			total += asset.DownloadCount
		}
	}

	type result struct {
		Total     int            `json:"total"`
		Platforms map[string]int `json:"platforms"`
	}

	sorted := make(map[string]int)
	keys := make([]string, 0, len(platforms))
	for k := range platforms {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return platforms[keys[i]] > platforms[keys[j]]
	})
	for _, k := range keys {
		sorted[k] = platforms[k]
	}

	out := result{Total: total, Platforms: sorted}

	outFile := filepath.Join("data", "stats.json")

	f, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "stats: %v\n", err)
		os.Exit(1)
	}
	f = append(f, '\n')

	if err := os.WriteFile(outFile, f, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "stats: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Updated %s: %d total downloads\n", outFile, total)
}
