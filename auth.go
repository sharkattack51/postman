package main

import (
	jwt "github.com/golang-jwt/jwt/v5"
)

func GenerateToken(scrt string, key string) (string, error) {
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"key": key,
	})
	tknStr, err := tkn.SignedString([]byte(scrt))
	if err != nil {
		return "", err
	}

	return tknStr, nil
}

func Authenticate(scrt string, tknStr string, key string) (bool, error) {
	tkn, err := jwt.Parse(tknStr, func(tkn *jwt.Token) (interface{}, error) {
		return []byte(scrt), nil
	})
	if err != nil {
		return false, err
	}

	claims := tkn.Claims.(jwt.MapClaims)
	return claims["key"].(string) == key, nil
}
