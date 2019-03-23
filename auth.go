package main

import (
	jwt "github.com/dgrijalva/jwt-go"
)

type Auth struct {
	Key string `json:"key"`
	jwt.StandardClaims
}

func GenerateToken(scrt string, key string) (string, error) {
	tkn := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), &Auth{
		Key: key,
	})
	token, err := tkn.SignedString([]byte(scrt))
	if err != nil {
		return "", err
	}

	return token, nil
}

func Authenticate(scrt string, token string, key string) (bool, error) {
	auth := Auth{}
	_, err := jwt.ParseWithClaims(token, &auth, func(tkn *jwt.Token) (interface{}, error) {
		return []byte(scrt), nil
	})
	if err != nil {
		return false, err
	}

	return auth.Key == key, nil
}
