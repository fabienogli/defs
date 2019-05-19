package services

import (
	jwt "github.com/dgrijalva/jwt-go"
)

//@todo /!\ Irindul : Get this from os.GetEnv(JWT_SECRET)
var TokenSecret = "mysupersecrettoken"

type Claims struct {
	UserId uint
	jwt.StandardClaims
}


