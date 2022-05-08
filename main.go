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

/*
Main Function:
1.Opens database and defers to close when it is done
2. Calls create table where it creates a games table if it is not already made with the required fields for the project
3. Calls add game data where it uses excelize to open the Excel sheet and fill the database row by row
4. Calls gin to make a self-hosted api that uses the url section after /search as a context param and throws that into
find game function
5. Find Game function uses the c context param from gin to do a select statement from the database and find the info
needed to display then returns it into the parameters set
6. gin converts to JSON and sends the info to the localhost @ port 8090(i.e. localhost:8090/search/Counter-Strike)

The user can use the url as a search bar to find the game of their choosing along as they use the proper Query Name that
the game is listed under in the Excel sheet. For example, localhost:8090/search/Counter-Strike will bring up the desired game
Counter-Strike but Counter Strike will "strike" an error. So please use caution when spelling and punctuating. Besides these
points above the program self-hosts and allows the user to search the 13,283 different listings from the Excel sheet and their
desired data.

*/
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

/*
Open Data Base Function:
Handles the job of creating the database.
*/
func OpenDataBase(dbfile string) *sql.DB {
	database, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
	return database
}

/*
Create Tables Function:
Handles the job of creating the tables for the database given there is no such table already in the databse.
*/
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

/*
Add Game Data Function:
Adds the data to the database from the Excel sheet using excelize and a INSERT statement. The INSERT statement includes
an or IGNORE to solve the problem of duplicates throwing us out of filling the database. We
*/
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

/*
Find Game Function:
Finds the game passed in the url after /search and returns the required info for parsing into JSON format by gin in the main
function by using a SELECT statement. The SELECT statement grabs all the info by using a * specifying the table using FROM GAMES and
the specific game using WHERE gameName = ?, game. The ? is subbed with game which is given to us as a parameter of the function, and
that is passed in main where it was grabbed by gin from the url.
*/
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
