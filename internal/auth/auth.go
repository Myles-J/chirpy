package auth

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hashed), err
}

func CheckPassword(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(now),
		Issuer:    "chirpy",
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSecret))
}

// ValidateJWT parses and validates a JWT, returning the userID if valid.
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	const leeway = 5 * time.Second
	// We will parse the token into a struct that includes or is RegisteredClaims
	// as that's what was used to create the token.
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key to validate the signature
		return []byte(tokenSecret), nil
	},
		// Add leeway to account for potential clock drift
		jwt.WithLeeway(leeway),
	)

	if err != nil {
		// Return the specific error from parsing
		return uuid.Nil, fmt.Errorf("failed to parse or validate JWT: %w", err)
	}

	// Check if the token is valid after parsing
	if !token.Valid {
		return uuid.Nil, errors.New("invalid JWT token")
	}

	// Extract the userID from the Subject claim
	userIDStr := claims.Subject
	if userIDStr == "" {
		return uuid.Nil, errors.New("JWT subject (userID) is missing or empty")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse userID from JWT subject: %w", err)
	}

	// Token is valid and userID is extracted
	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	slog.Info(authHeader)
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}

	return splitAuth[1], nil
}

// MakeRefreshToken makes a random 256 bit token
// encoded in hex.
func MakeRefreshToken() (string, error) {
	const tokenLength = 32
	token := make([]byte, tokenLength)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}
