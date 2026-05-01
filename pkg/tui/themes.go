// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: BSD-3-Clause

package tui

// theme provides the colors for the properties.
// you can use ANSI, ANSI256 or Hex colors.
// https://html-color.code
type theme struct {
	mainTitleColor           string
	selectedBorderColor      string
	selectedTitleColor       string
	selectedDescriptionColor string
}

var themes = map[string]theme{
	"matrix": matrixTheme(),
	"sky":    skyTheme(),
}

func matrixTheme() theme {
	return theme{
		mainTitleColor:           "#648c11",
		selectedTitleColor:       "#9efd38",
		selectedBorderColor:      "#9efd38",
		selectedDescriptionColor: "#648c11",
	}
}

func skyTheme() theme {
	return theme{
		mainTitleColor:           "#4682b4",
		selectedTitleColor:       "#00bfff",
		selectedBorderColor:      "#00bfff",
		selectedDescriptionColor: "#4682b4",
	}
}
