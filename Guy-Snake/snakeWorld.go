package main

import (
	"math"
	"time"

	"github.com/gammazero/deque"
)

const (
	NORTH Direction = 2
	EAST  Direction = 1
	SOUTH Direction = -2
	WEST  Direction = -1
)

type Point struct {
	x, y int
}
type Limb struct {
	point  Point
	d      Direction
	length int
}
type Snake struct {
	limbs deque.Deque[*Limb]
	score int
}
type Border struct {
	point          Point
	height, weight int
}
type TextBox struct {
	text  string
	bold  bool
	point Point
}
type World struct {
	border           Border
	snake            Snake
	apple            Point
	scoreTextBox     TextBox
	highScoreTextBox TextBox
	delay            time.Duration
}

func newSnake(head *Point, d Direction, size int) *Snake {
	snake := new(Snake)
	snake.limbs.PushBack(&Limb{point: Point{head.x, head.y}, d: -d, length: size})
	snake.score = 0
	return snake
}

func getNextPoint(p *Point, d Direction, distance int) *Point {
	switch d {
	case NORTH:
		return &Point{p.x, p.y - distance}
	case EAST:
		return &Point{p.x + distance, p.y}
	case SOUTH:
		return &Point{p.x, p.y + distance}
	case WEST:
		return &Point{p.x - distance, p.y}
	}
	return &Point{0, 0}
}
func samePoint(p1, p2 *Point) bool {
	return p1.x == p2.x && p1.y == p2.y
}

func (limb *Limb) getLimbEndPoint() Point {
	return *getNextPoint(&limb.point, limb.d, limb.length)
}
func checkBetweenNumbers(a, b, check int) bool {
	max := math.Max(float64(a), float64(b))
	min := math.Min(float64(a), float64(b))
	return check >= int(min) && check <= int(max)
}
func (limb *Limb) checkPointInLimb(p *Point) bool {
	p1 := limb.point
	p2 := limb.getLimbEndPoint()
	return checkBetweenNumbers(p1.x, p2.x, p.x) && checkBetweenNumbers(p1.y, p2.y, p.y)
}
func (snake *Snake) checkPointOnSnake(p *Point) bool {
	for i := 0; i < snake.limbs.Len(); i++ {
		check := snake.limbs.At(i).checkPointInLimb(p)
		if check {
			return true
		}
	}
	return false
}
func (border *Border) checkPointOnBorder(p *Point) bool {
	return p.x == border.point.x || p.x == border.point.x+border.weight ||
		p.y == border.point.y || p.y == border.point.y+border.height
}

func (snake *Snake) changeSnakeDirection(newD Direction) {
	headLimb := snake.limbs.Front()
	if headLimb.d == newD || headLimb.d == -newD || headLimb.length == 0 {
		return
	}
	snake.limbs.PushFront(&Limb{point: Point{headLimb.point.x, headLimb.point.y}, d: -newD, length: 0})

}
func (textBox *TextBox) checkPointInTextBox(p Point) bool {
	row := textBox.point.y
	col := textBox.point.x
	for _, r := range []rune(textBox.text) {
		if row == p.y && col == p.x {
			return true
		}
		col++
		if r == '\n' {
			row++
			col = textBox.point.x
		}
	}
	return false
}
