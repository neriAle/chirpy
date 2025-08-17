package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:		"chirpy",
		IssuedAt:	jwt.NewNumericDate(time.Now()),
		ExpiresAt:	jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:	userID.String(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
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

func GetBearerToken(headers http.Header) (string, error) {
	authInfo := headers.Get("Authorization")
	if authInfo == "" {
		return "", errors.New("The header doesn't contain a token")
	}

	token := strings.Fields(authInfo)[1]
	return token, nil
}

func MakeRefreshToken() (string, error) {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}

	tokenString := hex.EncodeToString(tokenBytes)
	return tokenString, nil
}