package main

import (
	//"fmt"

	"github.com/gdamore/tcell/v2"
)

var Style = tcell.StyleDefault

func drawTextBox(s tcell.Screen, textBox *TextBox) {
	row := textBox.point.y
	col := textBox.point.x
	for _, r := range []rune(textBox.text) {
		s.SetContent(col, row, r, nil, Style)
		col++
		if r == '\n' {
			row++
			col = textBox.point.x
		}
	}
}

func drawLine(s tcell.Screen, r rune, point *Point, d Direction, length int) {
	var dx, dy int
	switch d {
	case NORTH:
		dx = 0
		dy = -1
	case EAST:
		dx = 1
		dy = 0
	case SOUTH:
		dx = 0
		dy = 1
	case WEST:
		dx = -1
		dy = 0
	}
	for i := 0; i < length; i++ {
		s.SetContent(point.x+i*dx, point.y+i*dy, r, nil, Style)
	}
}
func drawBorder(s tcell.Screen, border *Border) {
	p1 := &border.point
	p2 := &Point{border.point.x + border.weight, border.point.y + border.height}
	drawLine(s, tcell.RuneVLine, p1, SOUTH, border.height)
	drawLine(s, tcell.RuneVLine, p2, NORTH, border.height)
	drawLine(s, tcell.RuneHLine, p1, EAST, border.weight)
	drawLine(s, tcell.RuneHLine, p2, WEST, border.weight)
	s.SetContent(p1.x, p1.y, tcell.RuneULCorner, nil, Style)
	s.SetContent(p2.x, p1.y, tcell.RuneURCorner, nil, Style)
	s.SetContent(p1.x, p2.y, tcell.RuneLLCorner, nil, Style)
	s.SetContent(p2.x, p2.y, tcell.RuneLRCorner, nil, Style)
}

func drawSnake(s tcell.Screen, snake *Snake) {

	for i := 0; i < snake.limbs.Len(); i++ {
		iLimb := snake.limbs.At(i)
		drawLine(s, tcell.RuneBlock, &iLimb.point, iLimb.d, iLimb.length)
	}
}

func drawWorld(s tcell.Screen, world *World) {
	s.SetContent(world.apple.x, world.apple.y, tcell.RuneDiamond, nil, Style)
	drawSnake(s, &world.snake)
	drawBorder(s, &world.border)
	drawTextBox(s, &world.scoreTextBox)
	drawTextBox(s, &world.highScoreTextBox)
}
func drawMainMenu(s tcell.Screen, menu *MainMenu) {
	drawBorder(s, &menu.border)
	drawTextBox(s, &menu.title)
	drawTextBox(s, &menu.newGame)
	drawTextBox(s, &menu.settings)
	drawTextBox(s, &menu.scoreBoard)
}
