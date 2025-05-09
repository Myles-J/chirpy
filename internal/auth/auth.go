package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hashed), err

}

func CheckPassword(hash, password string) (error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}
