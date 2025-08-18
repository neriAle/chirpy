package auth

import(
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authInfo := headers.Get("Authorization")
	if authInfo == "" {
		return "", errors.New("Missing API key")
	}

	apiKey := strings.Fields(authInfo)[1]
	return apiKey, nil
}