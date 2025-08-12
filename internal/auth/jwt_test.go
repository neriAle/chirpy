package auth

import(
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestJWTValidSubject(t *testing.T) {
	tokenSecret := "testing"
	const expirationTime = time.Second

	id := uuid.New()
	jwt, err := MakeJWT(id, tokenSecret, expirationTime)
	if err != nil {
		t.Errorf("Error generating the JWT: %v", err)
		return
	}
	validatedId, err := ValidateJWT(jwt, tokenSecret)
	if err != nil {
		t.Errorf("Error validating the JWT: %v", err)
		return
	}
	if validatedId != id {
		t.Errorf("JWT subject is incorrect")
		return
	}
}

func TestJWTExpiration(t *testing.T) {
	tokenSecret := "testingExp"
	const expirationTime = time.Second
	const waitTime = expirationTime + 5 * time.Millisecond
	
	jwt, err := MakeJWT(uuid.New(), tokenSecret, expirationTime)
	if err != nil {
		t.Errorf("Error generating the JWT: %v", err)
		return
	}
	_, err = ValidateJWT(jwt, tokenSecret)
	if err != nil {
		t.Errorf("Error validating the JWT: %v", err)
		return
	}

	time.Sleep(waitTime)

	_, err = ValidateJWT(jwt, tokenSecret)
	if err == nil {
		t.Errorf("Expected token to be already expired")
		return
	}
}

func TestJWTWrongSecret(t *testing.T) {
	tokenSecret := "testing"
	const expirationTime = time.Second

	jwt, err := MakeJWT(uuid.New(), tokenSecret, expirationTime)
	if err != nil {
		t.Errorf("Error generating the JWT: %v", err)
		return
	}
	_, err = ValidateJWT(jwt, "wrongSecret")
	if err == nil {
		t.Errorf("Expected JWT to be invalid")
	}
}