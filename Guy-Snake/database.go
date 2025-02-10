package main

import (
	"fmt"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

func initDataBase() *sqinn.Sqinn {
	database := sqinn.MustLaunch(sqinn.Options{})
	database.MustOpen("./SnakeDB.db")
	database.ExecOne("CREATE TABLE SCORE_BOARD (score INTEGER,name VARCHAR )")
	database.ExecOne("CREATE TABLE CURRENT_GAME (gameID INTEGER,score INTEGER,difficulty INTEGER,size INTEGER )")
	initCurrentGame(database)
	//database.MustExecOne(fmt.Sprintf("INSERT INTO SCORE_BOARD (score) VALUES (%d)", 0))
	return database
}

func getHighScore(dataBase *sqinn.Sqinn) int {
	rows, err := dataBase.Query("SELECT score FROM SCORE_BOARD ORDER BY score DESC", nil, []byte{sqinn.ValInt})
	if err == nil {
		return 0
	}
	return rows[0].Values[0].AsInt()
}
func getCurrentGameScore(dataBase *sqinn.Sqinn) int {
	rows := dataBase.MustQuery("SELECT score FROM CURRENT_GAME", nil, []byte{sqinn.ValInt})
	return rows[0].Values[0].AsInt()
}
func setCurrentGameSettings(dataBase *sqinn.Sqinn, difficulty int, size int) {
	dataBase.MustExecOne(fmt.Sprintf("UPDATE CURRENT_GAME SET difficulty = %d, size = %d WHERE gameID = 1", difficulty, size))
}
func getCurrentGameSettings(dataBase *sqinn.Sqinn) (difficulty, size int) {
	rows := dataBase.MustQuery("SELECT difficulty,size FROM CURRENT_GAME WHERE gameID = 1", nil, []byte{sqinn.ValInt, sqinn.ValInt})
	return rows[0].Values[0].AsInt(), rows[0].Values[1].AsInt()
}
func setCurrentGameScore(dataBase *sqinn.Sqinn, currentScore int) {
	dataBase.MustExecOne(fmt.Sprintf("UPDATE CURRENT_GAME SET score = %d WHERE gameID = 1", currentScore))
}
func initCurrentGame(dataBase *sqinn.Sqinn) {
	_, err := dataBase.Query("SELECT gameID FROM CURRENT_GAME", nil, []byte{sqinn.ValInt})
	if err == nil {
		dataBase.MustExecOne("INSERT INTO CURRENT_GAME (gameID,score,difficulty,size) VALUES(1,0,1,1)")
	}
}
func registerGameToScoreBoard(dataBase *sqinn.Sqinn, name string) {
	dataBase.MustExecOne(fmt.Sprintf("INSERT INTO SCORE_BOARD (score,name) VALUES((SELECT score from CURRENT_GAME),'%s')", name))
}
