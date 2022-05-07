package main

import (
	"database/sql"
	_ "fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xuri/excelize/v2"
	"log"
	"strconv"
	_ "strings"
)

func main() {
	myDatabase := OpenDataBase("./Demo.db")
	defer myDatabase.Close()
	create_tables(myDatabase)
	addGameData(myDatabase)

	r := gin.Default()
	r.GET("/search/:gamename", func(c *gin.Context) {
		gameName, owners, metaCritic, recomm, releasedate, reqAge, systems, playEst := findGame(myDatabase, c.Param("gamename"))
		c.JSON(200, gin.H{
			"Game Name":        gameName,
			"Game Owners":      owners,
			"MetaCritic Score": metaCritic,
			"Recommendations":  recomm,
			"Release Date":     releasedate,
			"Required Age":     reqAge,
			"Systems":          systems,
			"Player Estimate":  playEst,
		})

	})
	r.Run(":8090") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func OpenDataBase(dbfile string) *sql.DB {
	database, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
	return database
}

func create_tables(database *sql.DB) {
	createStatement1 := "CREATE TABLE IF NOT EXISTS GAMES(    " +
		"game_Name TEXT NOT NULL PRIMARY KEY," +
		"ownersCount INTEGER," +
		"metaCriticScore INTEGER," +
		"recommendations INTEGER," +
		"releaseDate TEXT," +
		"requiredAge INTEGER," +
		"systems TEXT," +
		"playerEstimate INTEGER);"
	_, err := database.Exec(createStatement1)
	if err != nil {
		return
	}
}

func addGameData(database *sql.DB) {
	insert_statement := "INSERT or IGNORE INTO GAMES (game_Name, ownersCount, metaCriticScore, recommendations, releaseDate, requiredAge, systems, playerEstimate) VALUES (?,?,?,?,?,?,?,?);"
	excelFile, err := excelize.OpenFile("games-features (1).xlsx")
	if err != nil {
		log.Fatalln(err)
	}

	systems := ""
	all_rows, err := excelFile.GetRows("games-features")
	if err != nil {
		log.Fatalln(err)
	}

	for _, row := range all_rows {
		gameName := row[2]
		ownersCount, _ := strconv.Atoi(row[15])
		metaCriticScore, _ := strconv.Atoi(row[9])
		recommendations, _ := strconv.Atoi(row[12])
		releaseDate := row[4]
		requiredAge, _ := strconv.Atoi(row[5])
		systempc := row[26]
		if systempc == "True" {
			systems = systems + "pc "
		}
		sysetmlinux := row[27]
		if sysetmlinux == "True" {
			systems = systems + "linux "
		}
		systemmac := row[28]
		if systemmac == "True" {
			systems = systems + "mac"
		}
		playerEstimate, _ := strconv.Atoi(row[17])

		prepped_statement, err := database.Prepare(insert_statement)
		if err != nil {
			log.Fatalln(err)
		}
		_, err = prepped_statement.Exec(gameName, ownersCount, metaCriticScore, recommendations, releaseDate, requiredAge, systems, playerEstimate)
		if err != nil {
			return
		}
		systems = ""
	}
}

func findGame(database *sql.DB, game string) (string, int, int, int, string, int, string, int) {
	var gameName, releasedate, systems string
	var owners, metaCritic, recomm, reqAge, playEst int

	resultset, err := database.Query("SELECT * FROM GAMES WHERE game_Name = ?", game)
	if err != nil {
		log.Fatalln("couldn't get data: ", err)
	}
	defer resultset.Close()

	for resultset.Next() {
		err = resultset.Scan(&gameName, &owners, &metaCritic, &recomm, &releasedate, &reqAge, &systems, &playEst)
		if err != nil {
			log.Fatalln("error reading data from resultset", err)
		}
	}
	return gameName, owners, metaCritic, recomm, releasedate, reqAge, systems, playEst
}
