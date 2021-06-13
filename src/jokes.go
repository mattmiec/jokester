package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Joke contains information about a single Joke
type Joke struct {
	ID    int    `json:"id" binding:"required"`
	Likes int    `json:"likes"`
	Joke  string `json:"joke" binding: "required"`
}

const (
	host      = "localhost"
	port      = 5432
	user      = "postgres"
	password  = "password"
	dbname    = "jokeish"
	tablename = "jokes"
)

func DbConn() (*sql.DB, error) {
	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)

	if err != nil {
		return nil, err
	}

	// check db
	err = db.Ping()

	if err != nil {
		return nil, err
	}

	fmt.Println("Connected!")
	return db, err
}

func NewJokeDb(db *sql.DB, joke string) error {

	newJokeStmt := `insert into "jokes"("joke", "likes") values ($1, 0)`

	//joke1 := "Why did the chicken cross the road? \n\n To get to the other side!"
	_, err := db.Exec(newJokeStmt, joke)

	return err
}

func LikeJokeDb(db *sql.DB, id int) error {

	updateJokeStmt := `update "jokes" set "likes" = "likes" + 1 where "id"=$1`
	_, err := db.Exec(updateJokeStmt, id)

	return err
}

func ListJokesDb(db *sql.DB) ([]Joke, error) {

	rows, err := db.Query(`select "id", "joke", "likes" from "jokes"`)

	var jokes = make([]Joke, 0, 100)

	for rows.Next() {
		var id int
		var joke string
		var likes int

		err = rows.Scan(&id, &joke, &likes)
		if err != nil {
			return nil, err
		}

		jokestruct := Joke{
			ID:    id,
			Joke:  joke,
			Likes: likes,
		}

		jokes = append(jokes, jokestruct)
	}

	return jokes, err
}
