package jwt

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	Userid   string `json:"userid"`
	Password string `json:"password"`
	jwt.StandardClaims
}

// 编解码私钥，在生产环境中，该私钥请使用生成器生成，并妥善保管，此处使用简单字符串
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
