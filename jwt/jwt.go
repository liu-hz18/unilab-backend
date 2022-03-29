package jwt

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	userid   string `json:"userid"`
	password string `json:"password"`
	jwt.StandardClaims
}

var jwtSecret = []byte("hello unilab")

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

func TokenGenerator(userid, password string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(time.Hour * 6)
	claims := Claims{
		userid,
		password,
		jwt.StandardClaims{
			Audience:  userid,
			Id:        userid,
			ExpiresAt: expireTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "gin hello",
			NotBefore: time.Now().Unix(), // 生效时间
			Subject:   "login",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}
