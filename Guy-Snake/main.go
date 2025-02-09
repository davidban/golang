package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/cvilsmeier/sqinn-go/sqinn"
	"github.com/gdamore/tcell/v2"
)

type Direction int
type ScreenStatus int

const (
	MAIN_MENU ScreenStatus = iota
	WORLD
	INSERT_SCOREBOARD
	SCOREBOARD
	SETTINGS
	QUIT
)

type SnakeManager struct {
	screen       tcell.Screen
	screenStatus ScreenStatus
	dataBase     *sqinn.Sqinn
}

type MainMenu struct {
	border                               Border
	title, newGame, scoreBoard, settings TextBox
}

func (world *World) moveSnake(s tcell.Screen) bool {

	headLimb := world.snake.limbs.Front()
	tailLimb := world.snake.limbs.Back()
	newHeadPoint := getNextPoint(&headLimb.point, -headLimb.d, 1)
	tail := tailLimb.getLimbEndPoint()
	if world.border.checkPointOnBorder(newHeadPoint) || (world.snake.checkPointOnSnake(newHeadPoint) && !samePoint(newHeadPoint, &tail)) {
		return true
	}
	headLimb.point = *newHeadPoint
	headLimb.length++
	s.SetContent(headLimb.point.x, headLimb.point.y, tcell.RuneBlock, nil, Style)

	if world.apple.x == headLimb.point.x && world.apple.y == headLimb.point.y {
		world.snake.score++
		generateNewApple(s, world)
		world.scoreTextBox.text = fmt.Sprintf("%s%d", "SCORE: ", world.snake.score)
		drawTextBox(s, &world.scoreTextBox)
		return false
	}
	tailLimb.length--
	tail = tailLimb.getLimbEndPoint()
	s.SetContent(tail.x, tail.y, ' ', nil, Style)
	if tailLimb.length == 0 {
		world.snake.limbs.PopBack()
	}
	return false
}
func generateNewApple(s tcell.Screen, world *World) {
	var randX, randY int
	for {
		randX = rand.Intn(world.border.weight-2) + world.border.point.x + 1
		randY = rand.Intn(world.border.height-2) + world.border.point.y + 1
		if !(world.snake.checkPointOnSnake(&Point{randX, randY})) {
			break
		}
	}
	world.apple = Point{randX, randY}
	s.SetContent(world.apple.x, world.apple.y, tcell.RuneDiamond, nil, Style)
}
func initScreen() tcell.Screen {
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	s.SetStyle(defStyle)
	s.EnableMouse()
	s.Clear()
	return s
}
func initDataBase() *sqinn.Sqinn {
	database := sqinn.MustLaunch(sqinn.Options{})
	database.MustOpen("./SnakeDB.db")
	database.ExecOne("CREATE TABLE SCORE_BOARD (score INTEGER,name VARCHAR )")
	database.ExecOne("CREATE TABLE CURRENT_GAME (score INTEGER )")
	database.ExecOne("CREATE TABLE SETTINGS (color VARCHAR )")
	//database.MustExecOne(fmt.Sprintf("INSERT INTO SCORE_BOARD (score) VALUES (%d)", 0))
	return database
}
func runSnakeGame() {
	//init Screen
	screen := initScreen()
	database := initDataBase()
	quit := func() {
		maybePanic := recover()
		screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
		database.Terminate()
		database.Close()
	}
	defer quit()

	snakeManager := SnakeManager{
		screen:       screen,
		screenStatus: MAIN_MENU,
		dataBase:     database,
	}

	for {
		switch snakeManager.screenStatus {
		case MAIN_MENU:
			snakeManager.runMainMenu()
		case WORLD:
			snakeManager.runWorld()
		case INSERT_SCOREBOARD:
			snakeManager.runInsertScoreboard()
		case SCOREBOARD:
			snakeManager.runScoreboard()
		case QUIT:
			return
		}
	}
}
func getHighScore(dataBase *sqinn.Sqinn) int {
	rows, err := dataBase.Query("SELECT score FROM SCORE_BOARD ORDER BY score DESC", nil, []byte{sqinn.ValInt})
	if err == nil {
		return 0
	}
	return rows[0].Values[0].AsInt()
}
func getCurrentScore(dataBase *sqinn.Sqinn) int {
	rows := dataBase.MustQuery("SELECT score FROM CURRENT_GAME", nil, []byte{sqinn.ValInt})
	return rows[0].Values[0].AsInt()
}
func registerCurrentGame(dataBase *sqinn.Sqinn, currentScore int) {
	dataBase.MustExecOne(fmt.Sprintf("INSERT INTO CURRENT_GAME (score) VALUES (%d)", currentScore))
}
func deleteCurrentGame(dataBase *sqinn.Sqinn) {
	dataBase.MustExecOne("DELETE FROM CURRENT_GAME")
}
func registerGameToScoreBoard(dataBase *sqinn.Sqinn, name string) {
	dataBase.MustExecOne(fmt.Sprintf("INSERT INTO SCORE_BOARD (score,name) VALUES((SELECT score from CURRENT_GAME),'%s')", name))
	deleteCurrentGame(dataBase)
}
func (world *World) snakeMoving(snakeManager *SnakeManager, wg *sync.WaitGroup, quit chan struct{}) {
	screen := snakeManager.screen
	for {
		select {
		case <-quit:
			return
		default:
			wg.Add(1)
			time.Sleep(world.delay)
			endGame := world.moveSnake(screen)
			screen.Show()
			wg.Done()
			if endGame {
				registerCurrentGame(snakeManager.dataBase, world.snake.score)
				snakeManager.screenStatus = INSERT_SCOREBOARD
				close(quit)
			}
		}

	}
}

func (snakeManager *SnakeManager) runWorld() {
	screen := snakeManager.screen
	screen.Clear()
	world := World{
		border:           Border{point: Point{0, 0}, height: 30, weight: 30},
		snake:            *newSnake(&Point{10, 10}, NORTH, 7),
		scoreTextBox:     TextBox{text: "SCORE: 0", point: Point{1, 31}},
		highScoreTextBox: TextBox{text: fmt.Sprintf("%s%d", "HIGH SCORE: ", getHighScore(snakeManager.dataBase)), point: Point{1, 32}},
		delay:            time.Millisecond * 80}
	generateNewApple(screen, &world)
	drawWorld(screen, &world)
	screen.Show()
	///
	var wg sync.WaitGroup
	evChannel := make(chan tcell.Event)
	quitChannel := make(chan struct{})
	go screen.ChannelEvents(evChannel, quitChannel)
	go world.snakeMoving(snakeManager, &wg, quitChannel)
	for {
		select {
		case <-quitChannel:
			wg.Wait()
			return
		case ev := <-evChannel:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				screen.Sync()
			case *tcell.EventKey:
				if ev.Rune() == 'r' || ev.Rune() == 'R' {
					close(quitChannel)
				}
				switch ev.Key() {
				case tcell.KeyCtrlC:
					snakeManager.screenStatus = QUIT
					close(quitChannel)
				case tcell.KeyEscape:
					snakeManager.screenStatus = MAIN_MENU
					close(quitChannel)
				case tcell.KeyUp:
					world.snake.changeSnakeDirection(NORTH)
				case tcell.KeyDown:
					world.snake.changeSnakeDirection(SOUTH)
				case tcell.KeyRight:
					world.snake.changeSnakeDirection(EAST)
				case tcell.KeyLeft:
					world.snake.changeSnakeDirection(WEST)
				}
			}
		}
	}
}

func (snakeManager *SnakeManager) runMainMenu() {
	s := snakeManager.screen
	s.Clear()
	mainMenu := MainMenu{
		border:     Border{point: Point{0, 0}, height: 30, weight: 30},
		title:      TextBox{point: Point{10, 10}, text: "SSSnake"},
		newGame:    TextBox{point: Point{10, 14}, text: "New Game"},
		scoreBoard: TextBox{point: Point{10, 16}, text: "Score Board"},
		settings:   TextBox{point: Point{10, 18}, text: "Settings"},
	}
	drawMainMenu(s, &mainMenu)

	s.Show()
	///
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				snakeManager.screenStatus = QUIT
				return
			}
		case *tcell.EventMouse:
			switch ev.Buttons() {
			case tcell.Button1:
				x, y := ev.Position()
				clickedPoint := Point{x, y}
				if mainMenu.newGame.checkPointInTextBox(clickedPoint) {
					snakeManager.screenStatus = WORLD
					return
				}
				if mainMenu.scoreBoard.checkPointInTextBox(clickedPoint) {
					snakeManager.screenStatus = SCOREBOARD
					return
				}
				if mainMenu.settings.checkPointInTextBox(clickedPoint) {
					snakeManager.screenStatus = SETTINGS
					return
				}
			}
		}

	}
}
func (snakeManager *SnakeManager) runInsertScoreboard() {
	screen := snakeManager.screen
	screen.Clear()
	border := Border{point: Point{0, 0}, height: 30, weight: 30}
	info := [2]TextBox{
		TextBox{point: Point{8, 10}, text: fmt.Sprintf("Your Score: %d", getCurrentScore(snakeManager.dataBase))},
		TextBox{point: Point{8, 11}, text: "Enter Your Name:"},
	}
	for _, tBox := range info {
		drawTextBox(screen, &tBox)
	}
	drawBorder(screen, &border)
	screen.Show()
	////
	var name string
	cursorLocation := Point{8, 12}
	screen.ShowCursor(cursorLocation.x, cursorLocation.y)
	defer screen.HideCursor()
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEsc:
				snakeManager.screenStatus = MAIN_MENU
				deleteCurrentGame(snakeManager.dataBase)
				return
			case tcell.KeyEnter:
				if name != "" {
					registerGameToScoreBoard(snakeManager.dataBase, name)
					snakeManager.screenStatus = MAIN_MENU
					return
				}
			case tcell.KeyBackspace:
				if len(name) > 0 {
					name = name[:len(name)-1]
					cursorLocation.x--
					screen.SetContent(cursorLocation.x, cursorLocation.y, 0, nil, Style)
					screen.ShowCursor(cursorLocation.x, cursorLocation.y)
					screen.Show()
				}
			case tcell.KeyRune:
				if len(name) < 10 {
					char := ev.Rune()
					name = fmt.Sprintf("%s%c", name, char)
					screen.SetContent(cursorLocation.x, cursorLocation.y, char, nil, Style)

					cursorLocation.x++
					screen.ShowCursor(cursorLocation.x, cursorLocation.y)
					screen.Show()
				}

			}
		}
	}
}
func drawScoreBoard(screen tcell.Screen, dataBase *sqinn.Sqinn) {
	rows := dataBase.MustQuery("SELECT score,name FROM SCORE_BOARD ORDER BY score DESC LIMIT 10", nil, []byte{sqinn.ValInt, sqinn.ValText})
	scoreText := TextBox{point: Point{5, 9}, text: "SCORE"}
	nameText := TextBox{point: Point{14, 9}, text: "NAME"}
	drawTextBox(screen, &scoreText)
	drawTextBox(screen, &nameText)

	for i, row := range rows {
		drawTextBox(screen, &TextBox{point: *getNextPoint(&scoreText.point, SOUTH, i+2), text: fmt.Sprintf("%d", row.Values[0].AsInt())})
		drawTextBox(screen, &TextBox{point: *getNextPoint(&nameText.point, SOUTH, i+2), text: row.Values[1].AsString()})
	}
}
func (snakeManager *SnakeManager) runScoreboard() {
	screen := snakeManager.screen
	screen.Clear()
	border := Border{point: Point{0, 0}, height: 30, weight: 30}
	drawBorder(screen, &border)
	drawScoreBoard(screen, snakeManager.dataBase)
	screen.Show()
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEsc {
				snakeManager.screenStatus = MAIN_MENU
				return
			}
		}
	}
}
func main() {

	runSnakeGame()

}
