package auth

import (
	"errors"
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

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	var id uuid.UUID
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return id, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return id, errors.New("Unknown or invalid claims")
	}

	id, err = uuid.Parse(claims.Subject)
	if err != nil {
		return id, errors.New("Invalid subject, can't be parsed into a uuid")
	}
	
	return id, nil
}