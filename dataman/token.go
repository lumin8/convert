package main

// courtesy vdparikh https://gist.github.com/vdparikh/bbaf5d023c65a3fbad4dc6b1e79497e0

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

const secret = "@Days3and2"

// User ...
// Custom object which can be stored in the claims
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthToken ...
// This is what is retured to the user
type AuthToken struct {
	TokenType string `json:"token_type"`
	Token     string `json:"access_token"`
	ExpiresIn int64  `json:"expires_in"`
}

// AuthTokenClaim ...
// This is the cliam object which gets parsed from the authorization header
type AuthTokenClaim struct {
	*jwt.StandardClaims
	User
}

// ErrorMsg ...
// Custom error object
type ErrorMsg struct {
	Message string `json:"message"`
}

func validate(w http.ResponseWriter, req *http.Request) {
	authorizationHeader := req.Header.Get("authorization")
	if authorizationHeader == "" {
		json.NewEncoder(w).Encode(ErrorMsg{Message: "An authorization header is required"})
		return
	}

	bearerToken := strings.Split(authorizationHeader, " ")
	if len(bearerToken) != 2 {
		json.NewEncoder(w).Encode(ErrorMsg{Message: "Invalid authorization token"})
		return
	}

	token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}
		return []byte(secret), nil
	})

	if err != nil {
		json.NewEncoder(w).Encode(ErrorMsg{Message: err.Error()})
		return
	}

	if !token.Valid {
		json.NewEncoder(w).Encode(ErrorMsg{Message: "Invalid authorization token"})
		return
	}

	var user User
	mapstructure.Decode(token.Claims, &user)
	vars := mux.Vars(req)

	name := vars["userId"]
	if name != user.Username {
		json.NewEncoder(w).Encode(ErrorMsg{Message: "Invalid authorization token - Does not match UserID"})
		return
	}

	context.Set(req, "decoded", token.Claims)
	//FIXME: What do we want to write here
	//w.Write(w, req)
}

func users(w http.ResponseWriter, req *http.Request) {
	decoded := context.Get(req, "decoded")
	var user User
	mapstructure.Decode(decoded.(jwt.MapClaims), &user)
	json.NewEncoder(w).Encode(user)
}

//call from main like thus
//    router.HandleFunc("/token/users/{userId}/credentials", validate(users)).Methods("GET")
