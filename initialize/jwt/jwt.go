package jwt

import (
	"ding/initialize/viper"
	"ding/model/dingding"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var MySecret = []byte("夏天夏天悄悄过去")

type MyClaims struct {
	UserId      string `json:"user_id"`
	Username    string `json:"user_name"`
	AuthorityID uint   `json:"authority_id"`
	jwt.StandardClaims
}

// GenToken 生成JWT
func GenToken(c *gin.Context, user *dingding.DingUser) (string, error) {
	fmt.Println(viper.Conf.Auth.Jwt_Expire)
	// 创建一个我们自己的声明
	m := MyClaims{
		user.UserId, // 自定义字段
		user.Name,
		user.AuthorityId,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(
				time.Duration(viper.Conf.Auth.Jwt_Expire) * time.Hour).Unix(), // 过期时间8760
			Issuer: "yjp", // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, m)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(MySecret)
}

// ParseToken 解析JWT
func (mc *MyClaims) ParseToken(tokenString string) (*MyClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, mc, func(token *jwt.Token) (i interface{}, err error) {
		return MySecret, nil
	})
	if err != nil {
		return nil, err
	}
	if token.Valid { // 校验token
		return mc, nil
	}
	return nil, errors.New("invalid token")
}
