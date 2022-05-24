package auth

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"time"
)

type Users struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type CustomClaims struct {
	UserName string
	jwt.StandardClaims
}

var Mysecret = []byte("mysecretxxx")

func GenToken(username string) (string, error) {
	claim := CustomClaims{
		UserName: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 5).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString(Mysecret)
}

func ParseToken(tokenStr string) (*CustomClaims, error) {

	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return Mysecret, nil
	})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func CheckUserInfo(claims *CustomClaims) bool {
	if claims.UserName == "admin" {
		return true
	}
	return false
}
func Verify(username, password string) bool {
	if username == "admin" && password == "admin" {
		return true
	}
	fmt.Println(username, password)
	return false
}
