package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NewJokeContainer struct {
	Joke string `json:"joke"`
}

type Joke struct {
	Joke      string    `json:"joke"`
	Author    string    `json:"author"`
	Likes     int       `json:"likes"`
	CreatedAt time.Time `json:"createdAt"`
}

func main() {

	// setup jwt middleware
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			aud := os.Getenv("AUTH0_API_AUDIENCE")
			checkAudience := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
			if !checkAudience {
				return token, errors.New("Invalid audience.")
			}
			// verify iss claim
			iss := os.Getenv("AUTH0_DOMAIN")
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				return token, errors.New("Invalid issuer.")
			}

			cert, err := getPemCert(token)
			if err != nil {
				log.Fatalf("could not get cert: %+v", err)
			}

			result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			return result, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
		UserProperty:  "jwt",
	})

	// register our actual jwtMiddleware
	jwtMiddleWare = jwtMiddleware

	// Set the router as the Gin default
	router := gin.Default()

	// Serve frontend static files
	router.Use(static.Serve("/", static.LocalFile("../views", true)))

	// Setup route group for the API
	api := router.Group("/api")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}
	// Two routes
	// /jokes - retrieves a list of jokes for a user
	// / jokes/like/:jokeID - capture likes sent to a particular joke
	api.GET("/jokes", authMiddleware(), userHandler(), ListJokes)
	api.POST("/jokes/like/:jokeID", authMiddleware(), userHandler(), LikeJoke)
	api.DELETE("/jokes/:jokeID", authMiddleware(), userHandler(), DeleteJoke)
	api.POST("/jokes/new", authMiddleware(), userHandler(), NewJoke)

	// Start and run the server
	router.Run(":3000")
}

// ListJokes retrieves a list of available jokes
func ListJokes(c *gin.Context) {

	db, err := DbConn()

	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
		return
	}

	jokes, err := ListJokesDb(db)

	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
		return
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, jokes)
}

// LikeJoke increments the likes of a particular joke Item
func LikeJoke(c *gin.Context) {
	// confirm Joke ID sent is valid
	// remember to import the `strconv` package
	if jokeID, err := uuid.Parse(c.Param("jokeID")); err == nil {

		db, err := DbConn()

		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
			return
		}

		userID := uuid.New()
		err = LikeJokeDb(db, userID, jokeID)

		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
			return
		}

	} else {
		// Joke ID is invalid
		c.AbortWithStatus(http.StatusNotFound)
	}
}

func NewJoke(c *gin.Context) {

	var newJokeContainer NewJokeContainer
	err := c.BindJSON(&newJokeContainer)
	if err != nil {
		return
	}

	db, err := DbConn()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
		return
	}

	userID := uuid.New()
	err = NewJokeDb(db, newJokeContainer.Joke, userID)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
		return
	}
}

func DeleteJoke(c *gin.Context) {
	if _, err := strconv.Atoi(c.Param("jokeID")); err == nil {

		db, err := DbConn()

		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
			return
		}

		jokeID := uuid.New()
		err = DeleteJokeDb(db, jokeID)

		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
			return
		}

	} else {
		// Joke ID is invalid
		c.AbortWithStatus(http.StatusNotFound)
	}
}
