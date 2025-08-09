package auth

import(
	"testing"
)

func TestHashing(t *testing.T) {
	cases := []string{"password1", "VerySecurepw123!", "hellothere", "437b930db84b8079c2dd804a71936b5f"}
	for _, c := range cases {
		hash, err := HashPassword(c)
		if err != nil {
			t.Errorf("Failed to hash password: %s", c)
			continue
		}
		err = CheckPasswordHash(c, hash)
		if err != nil {
			t.Errorf("Failed to check password's hash: %s. %s", c, err)
		}
	}
}