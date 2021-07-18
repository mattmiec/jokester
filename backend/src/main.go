package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NewUserContainer struct {
	UserName string `json:"username"`
}

type NewJokeContainer struct {
	Joke string `json:"joke"`
}

type Joke struct {
	JokeID  uuid.UUID `json:"id"`
	Joke    string    `json:"joke"`
	Created time.Time `json:"created"`
	Author  string    `json:"author"`
	Likes   int       `json:"likes"`
	Liked   bool      `json:"liked"`
}

func main() {

	// setup jwt middleware
	jwtMiddleWare = jwtmiddleware.New(getMiddlewareOptions())

	// Set the router as the Gin default
	router := gin.Default()

	// Serve frontend static files
	router.Use(static.Serve("/", static.LocalFile("../../frontend/dist", true)))

	// CORS
	router.Use(CORSMiddleware())

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
	api.POST("/user", authMiddleware(), NewUser)
	api.DELETE("/user", authMiddleware(), verifyUser(), DeleteUser)
	api.GET("/jokes", authMiddleware(), verifyUser(), ListJokes)
	api.POST("/jokes/like/:jokeID", authMiddleware(), verifyUser(), LikeJoke)
	api.POST("/jokes/unlike/:jokeID", authMiddleware(), verifyUser(), UnlikeJoke)
	api.DELETE("/jokes/:jokeID", authMiddleware(), verifyUser(), DeleteJoke)
	api.POST("/jokes/new", authMiddleware(), verifyUser(), NewJoke)

	// Start and run the server
	router.Run(":3000")
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func NewUser(c *gin.Context) {
	token, ok := c.Request.Context().Value("jwt").(*jwt.Token)
	if !ok {
		c.Abort()
		c.String(http.StatusInternalServerError, "Error parsing jwt from context")
		return
	}

	subject, ok := token.Claims.(jwt.MapClaims)["sub"]
	if !ok {
		c.Abort()
		c.String(http.StatusInternalServerError, "Error parsing subject from jwt")
		return
	}

	db, err := DbConn()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
		return
	}

	var newUserContainer NewUserContainer

	err = c.BindJSON(&newUserContainer)
	if err != nil {
		c.String(http.StatusBadRequest, "Incorrect json body")
		return
	}

	err = NewUserDb(db, subject.(string), newUserContainer.UserName)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
		return
	}

	db.Close()
}

func DeleteUser(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.String(http.StatusInternalServerError, "Unable to retrieve userID from context")
		return
	}

	db, err := DbConn()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
		return
	}

	err = DeleteUserDb(db, userID.(uuid.UUID))
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
		return
	}

	db.Close()
}

// ListJokes retrieves a list of available jokes
func ListJokes(c *gin.Context) {

	db, err := DbConn()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
		return
	}

	userID, ok := c.Get("userID")
	if !ok {
		c.String(http.StatusInternalServerError, "Unable to retrieve userID from context")
		return
	}

	jokes, err := ListJokesDb(db, userID.(uuid.UUID))

	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
		return
	}

	db.Close()

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, jokes)
}

func LikeJoke(c *gin.Context) {
	// confirm Joke ID sent is valid
	// remember to import the `strconv` package
	if jokeID, err := uuid.Parse(c.Param("jokeID")); err == nil {

		db, err := DbConn()

		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
			return
		}

		userID, ok := c.Get("userID")
		if !ok {
			c.String(http.StatusInternalServerError, "Unable to retrieve userID from context")
			return
		}

		err = LikeJokeDb(db, userID.(uuid.UUID), jokeID)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
			return
		}

		db.Close()

	} else {
		// Joke ID is invalid
		c.AbortWithStatus(http.StatusNotFound)
	}
}

func UnlikeJoke(c *gin.Context) {
	// confirm Joke ID sent is valid
	// remember to import the `strconv` package
	if jokeID, err := uuid.Parse(c.Param("jokeID")); err == nil {

		db, err := DbConn()

		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
			return
		}

		userID, ok := c.Get("userID")
		if !ok {
			c.String(http.StatusInternalServerError, "Unable to retrieve userID from context")
			return
		}

		err = UnlikeJokeDb(db, userID.(uuid.UUID), jokeID)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err))
			return
		}

		db.Close()

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

	db.Close()
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

		db.Close()

	} else {
		// Joke ID is invalid
		c.AbortWithStatus(http.StatusNotFound)
	}
}
