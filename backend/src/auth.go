package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gin-gonic/gin"
)

type UserInfo struct {
	Sub            string
	Given_name     string
	Family_name    string
	Nickname       string
	Name           string
	Picture        string
	Locale         string
	Updated_at     time.Time
	Email          string
	Email_verified bool
}

var jwtMiddleWare *jwtmiddleware.JWTMiddleware

// Jwks stores a slice of JSON Web Keys
type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func getMiddlewareOptions() jwtmiddleware.Options {
	return jwtmiddleware.Options{
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
	}
}

func verifyUser() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		userID, err := getUserIdDb(db, subject.(string))

		if err != nil {
			fmt.Println(err)
			c.Abort()
			c.Writer.WriteHeader(http.StatusForbidden)
		}

		c.Set("userID", *userID)
	}
}

func getUserInfo(c *gin.Context) (*UserInfo, error) {

	token, ok := c.Request.Context().Value("jwt").(*jwt.Token)
	if !ok {
		return nil, errors.New("Error parsing jwt from context")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", os.Getenv("AUTH0_DOMAIN")+"userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token.Raw)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(bodyBytes))

	userinfo := UserInfo{}
	err = json.NewDecoder(resp.Body).Decode(&userinfo)
	if err != nil {
		return nil, err
	}

	fmt.Println(userinfo)
	return &userinfo, nil
}

// authMiddleware intercepts the requests, and check for a valid jwt token
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the client secret key
		err := jwtMiddleWare.CheckJWT(c.Writer, c.Request)
		if err != nil {
			// Token not found
			fmt.Println(err)
			c.Abort()
			c.Writer.WriteHeader(http.StatusUnauthorized)
			c.Writer.Write([]byte("Unauthorized"))
			return
		}
	}
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get(os.Getenv("AUTH0_DOMAIN") + ".well-known/jwks.json")
	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	x5c := jwks.Keys[0].X5c
	for k, v := range x5c {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + v + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		return cert, errors.New("unable to find appropriate key.")
	}

	return cert, nil
}
