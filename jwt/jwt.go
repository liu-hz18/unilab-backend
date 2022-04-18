package jwt

import (
	"time"
	"unilab-backend/setting"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	Userid   string `json:"userid"`
	UserName string `json:"username"`
	jwt.StandardClaims
}

// 编解码私钥，在生产环境中，该私钥请使用生成器生成，并妥善保管，此处使用简单字符串
var jwtSecret = []byte(setting.JwtSecret)

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

func TokenGenerator(userid, username string) (string, error) {
	nowTime := time.Now()
	// 6h 后失效
	expireTime := nowTime.Add(time.Hour * 6)
	claims := Claims{
		userid,
		username,
		jwt.StandardClaims{
			Audience:  userid,
			Id:        userid,
			ExpiresAt: expireTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "unilab hello",
			NotBefore: time.Now().Unix(), // 生效时间
			Subject:   "login",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}
