package serializers

import (
	"balancer/common"
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

func (serializer *jwtUserDataSerializer) Serialize(userData *common.UserData) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"identifier": userData.Identifier,
		"username":   userData.Username,
		"email":      userData.Email,
		"locale":     userData.Locale,
		"picture":    userData.Picture,
	})
	return token.SignedString([]byte(serializer.hmacSampleSecret))
}
