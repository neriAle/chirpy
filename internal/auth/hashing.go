package auth

import(
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	pwData := []byte(password)
	hashData, err := bcrypt.GenerateFromPassword(pwData, 5)
	if err != nil {
		return "", err
	}
	hash := string(hashData)
	return hash, nil
}

func CheckPasswordHash(password, hash string) error {
	pwData := []byte(password)
	hashData := []byte(hash)
	return bcrypt.CompareHashAndPassword(hashData, pwData)
}