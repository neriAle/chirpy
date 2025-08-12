package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:		"chirpy",
		IssuedAt:	jwt.NewNumericDate(time.Now()),
		ExpiresAt:	jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:	userID.String()
	}
	
	token := jwt.NewWithClaim(jwt.SigningMethodHS256, claims)
	signingKey := []byte(tokenSecret)
	signedString, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}
	return signedString, nil
}