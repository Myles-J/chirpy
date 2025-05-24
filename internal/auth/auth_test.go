package auth_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Myles-J/chirpy/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ... (HashPassword, CheckPassword, MakeJWT, ValidateJWT remain the same) ...

func TestHashPassword(t *testing.T) {
	password := "mysecretpassword"
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned an error: %v", err)
	}
	if hashedPassword == "" {
		t.Error("HashPassword returned an empty string")
	}
	if errBcrypt := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); errBcrypt != nil {
		t.Errorf("CheckPassword failed for a valid hash: %v", errBcrypt)
	}
}

func TestCheckPassword(t *testing.T) {
	validPassword := "securepassword123"
	validHash, err := auth.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("Failed to hash password for test cases: %v", err)
	}

	tests := []struct {
		name         string
		hash         string
		password     string
		expectedErr  error
		expectErr    bool
		errSubstring string
	}{
		{"Correct Password", validHash, validPassword, nil, false, ""},
		{"Incorrect Password", validHash, "wrongpassword", bcrypt.ErrMismatchedHashAndPassword, true, ""},
		{"Invalid Hash Format", "not-a-valid-bcrypt-hash", validPassword, nil, true, ""},
	}

	for _, tt := range tests {
		errCheck := auth.CheckPassword(tt.hash, tt.password)
		if !tt.expectErr {
			if errCheck != nil {
				t.Errorf("%s: Expected no error, but got %v", tt.name, errCheck)
			}
			continue
		}
		if errCheck == nil {
			t.Errorf("%s: Expected an error, but got nil", tt.name)
			continue
		}
		if tt.expectedErr != nil && !errors.Is(errCheck, tt.expectedErr) {
			t.Errorf("%s: Expected error %v, but got %v", tt.name, tt.expectedErr, errCheck)
		}
		if tt.errSubstring != "" && !strings.Contains(errCheck.Error(), tt.errSubstring) {
			t.Errorf("%s: Expected error message to contain '%s', but got '%v'", tt.name, tt.errSubstring, errCheck)
		}
	}
}

func TestMakeJWT(t *testing.T) {
	tests := []struct {
		name      string
		userID    uuid.UUID
		jwtSecret string
		expiresIn time.Duration
		expectErr bool
	}{
		{"Standard JWT", uuid.New(), "a-secure-secret-for-test-make-jwt-1", 15 * time.Minute, false},
		{"Different User ID", uuid.New(), "a-secure-secret-for-test-make-jwt-2", 5 * time.Hour, false},
	}

	for _, tt := range tests {
		tokenString, errJWT := auth.MakeJWT(tt.userID, tt.jwtSecret, tt.expiresIn)
		if tt.expectErr {
			if errJWT == nil {
				t.Errorf("%s: Expected an error, but got nil", tt.name)
			}
		} else {
			if errJWT != nil {
				t.Errorf("%s: Expected no error, but got %v", tt.name, errJWT)
			}
			if tokenString == "" {
				t.Errorf("%s: MakeJWT returned an empty token string", tt.name)
			}
		}
	}
}

func TestValidateJWT(t *testing.T) {
	jwtSecret := "my-secure-test-secret-at-least-32-bytes-for-validate"
	validUserID := uuid.New()
	validExpiresIn := 1 * time.Minute
	validToken, err := auth.MakeJWT(validUserID, jwtSecret, validExpiresIn)
	if err != nil {
		t.Fatalf("Failed to make a valid token for test cases: %v", err)
	}
	expiredToken, err := auth.MakeJWT(validUserID, jwtSecret, -1*time.Minute)
	if err != nil {
		t.Fatalf("Failed to make an expired token for test cases: %v", err)
	}
	malformedClaimsToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(validExpiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "chirpy",
		Subject:   "not-a-uuid",
	})
	malformedTokenString, err := malformedClaimsToken.SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("Failed to sign malformed claims token for test cases: %v", err)
	}
	_, expectedParseErr := uuid.Parse("not-a-uuid")

	tests := []struct {
		name           string
		tokenString    string
		jwtSecret      string
		expectedUserID uuid.UUID
		expectedErr    error
		expectErr      bool
		errSubstring   string
	}{
		{"Valid Token", validToken, jwtSecret, validUserID, nil, false, ""},
		{"Invalid Secret", validToken, "wrong-secret", uuid.Nil, nil, true, "signature"},
		{"Expired Token", expiredToken, jwtSecret, uuid.Nil, nil, true, "expired"},
		{"Token with Malformed Subject", malformedTokenString, jwtSecret, uuid.Nil, expectedParseErr, true, ""},
		{"Empty Token String", "", jwtSecret, uuid.Nil, nil, true, "segments"},
	}

	for _, tt := range tests {
		validatedUserID, errValidate := auth.ValidateJWT(tt.tokenString, tt.jwtSecret)

		if !tt.expectErr {
			if errValidate != nil {
				t.Errorf("%s: Expected no error, but got %v", tt.name, errValidate)
			}
			if validatedUserID != tt.expectedUserID {
				t.Errorf(
					"%s: ValidateJWT returned incorrect UserID. Expected %s, got %s",
					tt.name,
					tt.expectedUserID,
					validatedUserID,
				)
			}
			continue
		}

		if errValidate == nil {
			t.Errorf("%s: Expected an error, but got nil", tt.name)
		}
		if tt.expectedErr != nil && !errors.Is(errValidate, tt.expectedErr) {
			t.Errorf("%s: Expected error to contain %v, but got %v", tt.name, tt.expectedErr, errValidate)
		}
		if tt.errSubstring != "" && !strings.Contains(errValidate.Error(), tt.errSubstring) {
			t.Errorf("%s: Expected error message to contain '%s', but got '%v'", tt.name, tt.errSubstring, errValidate)
		}
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		headers       http.Header
		expectedToken string
		expectErr     bool
	}{
		{"No Authorization header", http.Header{}, "", true},
		{
			"Malformed Authorization header (no Bearer)",
			http.Header{"Authorization": []string{"Basic abcdef"}},
			"",
			true,
		},
		{"Valid Bearer token", http.Header{"Authorization": []string{"Bearer my-token"}}, "my-token", false},
	}

	for _, tt := range tests {
		token, errBearer := auth.GetBearerToken(tt.headers)
		if !tt.expectErr {
			if errBearer != nil {
				t.Errorf("%s: Expected no error, but got %v", tt.name, errBearer)
			}
			if token != tt.expectedToken {
				t.Errorf("%s: Expected token %s, but got %s", tt.name, tt.expectedToken, token)
			}
			continue
		}
		if errBearer == nil {
			t.Errorf("%s: Expected an error, but got nil", tt.name)
		}
		if token != tt.expectedToken {
			t.Errorf("%s: Expected token %s on error, but got %s", tt.name, tt.expectedToken, token)
		}
	}
}

func TestMakeRefreshToken(t *testing.T) {
	token, err := auth.MakeRefreshToken()
	if err != nil {
		t.Fatalf("Failed to make a refresh token: %v", err)
	}
	if len(token) != 64 {
		t.Errorf("Expected token length to be 64, but got %d", len(token))
	}
}
