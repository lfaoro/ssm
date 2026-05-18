// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

import (
	"image/color"
	"strconv"
)

// theme provides the colors for the properties.
// you can use ANSI, ANSI256 or Hex colors.
type theme struct {
	backgroundColor         string
	mainTitleColor          string
	selectedBorderColor     string
	selectedTitleColor      string
	selectedDescriptionColor string
}

var themes = map[string]theme{
	"matrix": matrixTheme(),
	"sky":    skyTheme(),
}

func matrixTheme() theme {
	return theme{
		backgroundColor:         "#000000",
		mainTitleColor:          "#648c11",
		selectedTitleColor:      "#9efd38",
		selectedBorderColor:     "#9efd38",
		selectedDescriptionColor: "#648c11",
	}
}

func skyTheme() theme {
	return theme{
		backgroundColor:         "#0d1117",
		mainTitleColor:          "#4682b4",
		selectedTitleColor:      "#00bfff",
		selectedBorderColor:     "#00bfff",
		selectedDescriptionColor: "#4682b4",
	}
}

func parseHexColor(hex string) color.RGBA {
	if len(hex) == 7 && hex[0] == '#' {
		r, _ := strconv.ParseUint(hex[1:3], 16, 8)
		g, _ := strconv.ParseUint(hex[3:5], 16, 8)
		b, _ := strconv.ParseUint(hex[5:7], 16, 8)
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xff}
	}
	return color.RGBA{A: 0xff}
}
