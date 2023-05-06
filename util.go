package main

import "github.com/gdamore/tcell/v2"

func writeStr(s tcell.Screen, x, y int, style tcell.Style, text string) {
	row, col := y, x
	width, _ := s.Size()
	for _, r := range []rune(text) {
		s.SetContent(col, row, r, nil, style)
		col++
		if col >= width {
			col = 0
			row++
		}
	}
}
