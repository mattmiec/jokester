package main

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

func NewUserDb(db *sql.DB, openIdSub string, username string) error {

	newUserStmt := `insert into "users"("openid_sub","username") values ($1,$2)`

	_, err := db.Exec(newUserStmt, openIdSub, username)

	return err

}

func DeleteUserDb(db *sql.DB, userID uuid.UUID) error {

	deleteUserStmt := `delete from "users" where user_id=$1`

	_, err := db.Exec(deleteUserStmt, userID)

	return err

}

func getUserIdDb(db *sql.DB, openIdSub string) (*uuid.UUID, error) {
	getUserIdStmt := `select "user_id" from "users" where openid_sub=$1`

	rows, err := db.Query(getUserIdStmt, openIdSub)

	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, errors.New("openIdSub not found")
	}

	var userID uuid.UUID

	err = rows.Scan(&userID)

	if err != nil {
		return nil, err
	}

	if rows.Next() {
		return nil, errors.New("multiple matches for openIdSub")
	}

	return &userID, nil
}

func NewJokeDb(db *sql.DB, joke string, authorID uuid.UUID) error {

	newJokeStmt := `insert into "jokes"("joke", "author_id") values ($1, $2)`

	_, err := db.Exec(newJokeStmt, joke, authorID)

	return err
}

func DeleteJokeDb(db *sql.DB, jokeID uuid.UUID) error {

	deleteJokeStmt := `delete from "jokes" where "joke_id"=$1`
	_, err := db.Exec(deleteJokeStmt, jokeID)

	return err
}

func LikeJokeDb(db *sql.DB, userID uuid.UUID, jokeID uuid.UUID) error {

	likeJokeStmt := `insert into "likes"("user_id", "joke_id") values ($1, $2)`
	_, err := db.Exec(likeJokeStmt, userID, jokeID)

	return err
}

func UnlikeJokeDb(db *sql.DB, userID uuid.UUID, jokeID uuid.UUID) error {

	unlikeJokeStmt := `delete from "likes" where "user_id"=$1 and "joke_id"=$2`
	_, err := db.Exec(unlikeJokeStmt, userID, jokeID)

	return err
}

func ListJokesDb(db *sql.DB, userID uuid.UUID) ([]Joke, error) {

	rows, err := db.Query(`select joke_id, joke, created, 
	                       (select username from users where users.user_id = jokes.author_id) as author,
						   (select count(*) from likes where likes.joke_id = jokes.joke_id) as likes,
						   case when exists (select * from likes where likes.joke_id = jokes.joke_id and likes.user_id = $1) then 'true' else 'false' end as liked
						   from jokes`, userID)

	if err != nil {
		return nil, err
	}

	var jokes = make([]Joke, 0, 100)

	for rows.Next() {
		var joke Joke

		err = rows.Scan(&joke.JokeID, &joke.Joke, &joke.Created, &joke.Author, &joke.Likes, &joke.Liked)
		if err != nil {
			return nil, err
		}

		jokes = append(jokes, joke)
	}

	return jokes, nil
}
