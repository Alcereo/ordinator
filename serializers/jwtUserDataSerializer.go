package serializers

import (
	"balancer/auth"
	"github.com/dgrijalva/jwt-go"
)

type jwtUserDataSerializer struct {
	hmacSampleSecret string
}

func NewJwtUserDataSerializer(hmacSampleSecret string) *jwtUserDataSerializer {
	return &jwtUserDataSerializer{
		hmacSampleSecret: hmacSampleSecret,
	}
}

func (serializer *jwtUserDataSerializer) Serialize(userData *auth.UserData) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"identifier": userData.Identifier,
		"username":   userData.Username,
		"email":      userData.Email,
		"locale":     userData.Locale,
		"picture":    userData.Picture,
	})
	return token.SignedString([]byte(serializer.hmacSampleSecret))
}
