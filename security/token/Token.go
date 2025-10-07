package token

import (
	"ecommerce/components/log"

	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type Claims struct {
	UserId   string `json:"UserId"`
	UserName string `json:"UserName"`
	Role     string `json:"role"`

	jwt.StandardClaims
}

func GenerateAuthToken(userId uuid.UUID, userName string, role string) (string, error) {
	expirationTime := time.Now().Add(55 * time.Minute)
	claims := &Claims{
		UserId:   userId.String(),
		UserName: userName,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret := viper.GetString("JWT_SECRET")

	if jwtSecret == "" {
		log.GetLogger().Print("JWT_SECRET not found in environment variables")
	}

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.GetLogger().Print(err)
		return "", err
	}
	return tokenString, nil
}
