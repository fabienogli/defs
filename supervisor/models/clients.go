package models

import (
	"log"
	"golang.org/x/crypto/bcrypt"
	s "supervisor/services"
	jwt "github.com/dgrijalva/jwt-go"
)

type Client struct {
	Id uint `json:"id"`
	Username string `json:"username"`
	Email string `json:"email"`
	Password string `json:"-"`
	Token string `json:"token"`
}


func (c *Client) Sign() {
	jwtClaims := &s.Claims{UserId: c.Id}
	unsignedToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), jwtClaims)
	tokenString, err := unsignedToken.SignedString([]byte(s.TokenSecret))
	if err != nil {
		log.Fatalf("error while creating signed token : %s\n", err.Error())
	}
	c.Token = tokenString
}

func NewClient(username string, password string) (Client, error) {
	hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedPassword := string(hashedBytes)

	//Generate ID
	client := Client{
		Id: 1,
		Username: username,
		Password: hashedPassword, 
	}

	//@todo Save Client into DB

	//Get signed token for newly created user
	client.Sign()
	return client, nil

}