package util

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
)

var jwtSecret = []byte("shudi")

type Claims struct {
	ID        string `json:"user_id"`
	UserName  string `json:"user_name"`
	Password  string `json:"password"`
	Authority int    `json:"authority"`
	jwt.StandardClaims
}

//签发token

func GenerateToken(id string, passWord string, userName string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(24 * time.Hour)
	claims := Claims{
		ID:       id,
		UserName: userName,
		Password: passWord,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			IssuedAt:  nowTime.Unix(),
			Issuer:    "MALL-TEST",
			Subject:   "user token",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}
func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}
func UUID() string {
	return uuid.NewV4().String()
}
