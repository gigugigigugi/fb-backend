package utils

import (
	"errors"
	"football-backend/common/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims 是被签出 Token 中携带的用户数据黑匣子
type JWTClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken 生成一个有效期通常为数十小时的短期凭证
func GenerateToken(userID uint) (string, error) {
	// 最好提取自系统 config (config.App.JWT.Secret)，默认给个强口令兜底
	secret := config.App.JWT.Secret
	if secret == "" {
		secret = "SuperSecretKeyForFootballApp" // 在真实生产中不应硬编码
	}

	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "football-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 读取用户的 Token 证明，验证其不曾被篡改并合法剥离出 UserID
func ParseToken(tokenString string) (*JWTClaims, error) {
	secret := config.App.JWT.Secret
	if secret == "" {
		secret = "SuperSecretKeyForFootballApp"
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
