package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil
	}
	return string(hashedPassword), nil
}

func CheckPasswordCorrectness(hashedPassword, nativePassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(nativePassword))
}
