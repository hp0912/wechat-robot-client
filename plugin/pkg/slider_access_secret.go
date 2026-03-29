package pkg

import (
	"time"
	"wechat-robot-client/vars"

	jwt "github.com/golang-jwt/jwt/v5"
)

// TokenClaims JWT Claims结构体
type TokenClaims struct {
	DeviceIDs []string `json:"device_ids"`
	Exp       int64    `json:"exp"`
	jwt.RegisteredClaims
}

// GenerateSliderAccessSecret 生成JWT token
// deviceIDs: 设备ID列表，如果为空则生成超级token（允许所有设备访问）
// expireDuration: token过期时间，如果为0则默认24小时
func GenerateSliderAccessSecret(deviceIDs []string, expireDuration time.Duration) (string, error) {
	if expireDuration == 0 {
		expireDuration = 24 * time.Hour // 默认24小时过期
	}

	now := time.Now()
	expireTime := now.Add(expireDuration)

	claims := TokenClaims{
		DeviceIDs: deviceIDs,
		Exp:       expireTime.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireTime),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(vars.SliderAccessKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
