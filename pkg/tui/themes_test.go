// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

import (
	"testing"
)

func TestThemes(t *testing.T) {
	t.Run("matrix theme exists", func(t *testing.T) {
		th, ok := themes["matrix"]
		if !ok {
			t.Fatal("matrix theme not found")
		}
		if th.mainTitleColor == "" {
			t.Error("mainTitleColor is empty")
		}
		if th.selectedTitleColor == "" {
			t.Error("selectedTitleColor is empty")
		}
		if th.selectedBorderColor == "" {
			t.Error("selectedBorderColor is empty")
		}
		if th.selectedDescriptionColor == "" {
			t.Error("selectedDescriptionColor is empty")
		}
	})

	t.Run("sky theme exists", func(t *testing.T) {
		th, ok := themes["sky"]
		if !ok {
			t.Fatal("sky theme not found")
		}
		if th.mainTitleColor == "" {
			t.Error("mainTitleColor is empty")
		}
		if th.selectedTitleColor == "" {
			t.Error("selectedTitleColor is empty")
		}
		if th.selectedBorderColor == "" {
			t.Error("selectedBorderColor is empty")
		}
		if th.selectedDescriptionColor == "" {
			t.Error("selectedDescriptionColor is empty")
		}
	})

	t.Run("matrix and sky have distinct colors", func(t *testing.T) {
		matrixTh := themes["matrix"]
		skyTh := themes["sky"]

		if matrixTh.mainTitleColor == skyTh.mainTitleColor {
			t.Error("matrix and sky mainTitleColor should differ")
		}
	})
}
