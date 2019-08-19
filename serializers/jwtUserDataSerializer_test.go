package serializers

import (
	"balancer/auth"
	"github.com/stretchr/testify/assert"
	"testing"
)

const expectedToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InNvbWVAbWFpbC5ydSIsImlkZW50aWZpZXIiOiJ2ZXJ5VmVyeXZlcnlWZXJ5TG9uZ2lkZW50aWZpZXIiLCJsb2NhbGUiOiJydSIsInBpY3R1cmUiOiJwaWN0dXJlLXVybCIsInVzZXJuYW1lIjoic29tZS1uYW1lIn0.AdYdmEO8RDGya2AOzyKDiU7M_1XLBO5pQjUdECz5oww"

func TestSerializeToken(t *testing.T) {
	serializer := NewJwtUserDataSerializer("smallSecret")

	result, err := serializer.Serialize(&auth.UserData{
		Username:   "some-name",
		Picture:    "picture-url",
		Locale:     "ru",
		Email:      "some@mail.ru",
		Identifier: "veryVeryveryVeryLongidentifier",
	})
	assert.Empty(t, err)

	assert.Equal(t,
		expectedToken,
		result,
	)
}
