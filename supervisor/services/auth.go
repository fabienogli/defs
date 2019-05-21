package services

import (
	jwt "github.com/dgrijalva/jwt-go"
	"os"
)

var TokenSecret = os.Getenv("JWT_SECRET")

type Claims struct {
	UserId uint
	jwt.StandardClaims
}


