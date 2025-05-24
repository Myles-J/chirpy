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

// TestHashPassword tests the HashPassword function.
// Keeping this simple as it's a single main scenario.
func TestHashPassword(t *testing.T) {
	password := "mysecretpassword"
	hashedPassword, err := auth.HashPassword(password)

	if err != nil {
		t.Fatalf("HashPassword returned an error: %v", err)
	}

	if hashedPassword == "" {
		t.Error("HashPassword returned an empty string")
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		t.Errorf("CheckPassword failed for a valid hash: %v", err)
	}
}

// TestCheckPassword tests the CheckPassword function using test cases.
func TestCheckPassword(t *testing.T) {
	// Generate a valid hash for testing
	validPassword := "securepassword123"
	validHash, err := auth.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("Failed to hash password for test cases: %v", err)
	}

	// Define a slice of test cases
	tests := []struct {
		name         string
		hash         string
		password     string
		expectedErr  error  // The specific error we expect, or nil for success
		expectErr    bool   // Whether we expect any error
		errSubstring string // Substring to check in the error message if expected
	}{
		{
			name:        "Correct Password",
			hash:        validHash,
			password:    validPassword,
			expectedErr: nil,
			expectErr:   false,
		},
		{
			name:        "Incorrect Password",
			hash:        validHash,
			password:    "wrongpassword",
			expectedErr: bcrypt.ErrMismatchedHashAndPassword,
			expectErr:   true,
		},
		// Note: Testing invalid hash formats with bcrypt.CompareHashAndPassword
		// can return different error types or messages depending on the exact
		// malformation. Checking for any error is often sufficient here.
		{
			name:         "Invalid Hash Format",
			hash:         "not-a-valid-bcrypt-hash",
			password:     validPassword,
			expectedErr:  nil, // We expect *an* error, but the exact type can vary
			expectErr:    true,
			errSubstring: "", // Not checking for a specific substring due to variability
		},
	}

	// Iterate over the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { // t.Run allows for subtests with names
			err := auth.CheckPassword(tt.hash, tt.password)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				} else {
					// If a specific error is expected, check with errors.Is
					if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
						t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
					}
					// If checking for a substring (like for invalid hash), check that
					if tt.errSubstring != "" && !strings.Contains(err.Error(), tt.errSubstring) {
						t.Errorf("Expected error message to contain '%s', but got '%v'", tt.errSubstring, err)
					}
				}
			} else { // Expecting no error
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
			}
		})
	}
}

// TestMakeJWT tests the MakeJWT function using test cases.
func TestMakeJWT(t *testing.T) {
	// Define a slice of test cases
	tests := []struct {
		name      string
		userID    uuid.UUID
		jwtSecret string
		expiresIn time.Duration
		expectErr bool
	}{
		{
			name:      "Standard JWT",
			userID:    uuid.New(),
			jwtSecret: "a-secure-secret-for-test-make-jwt-1",
			expiresIn: 15 * time.Minute,
			expectErr: false,
		},
		{
			name:      "Different User ID",
			userID:    uuid.New(),
			jwtSecret: "a-secure-secret-for-test-make-jwt-2",
			expiresIn: 5 * time.Hour,
			expectErr: false,
		},
		// You could add cases for invalid secret lengths if the library
		// had specific error returns for that, but it typically doesn't for HMAC
	}

	// Iterate over the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := auth.MakeJWT(tt.userID, tt.jwtSecret, tt.expiresIn)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
				// No need to check specific error for MakeJWT in these cases
			} else { // Expecting no error
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
				if tokenString == "" {
					t.Error("MakeJWT returned an empty token string")
				}
				// Optional: Validate the token here to ensure it's parsable,
				// but TestValidateJWT covers this more comprehensively.
			}
		})
	}
}

// TestValidateJWT tests the ValidateJWT function using test cases.
func TestValidateJWT(t *testing.T) {
	jwtSecret := "my-secure-test-secret-at-least-32-bytes-for-validate" // Strong secret

	// Generate a valid token for testing
	validUserID := uuid.New()
	validExpiresIn := 1 * time.Minute
	validToken, err := auth.MakeJWT(validUserID, jwtSecret, validExpiresIn)
	if err != nil {
		t.Fatalf("Failed to make a valid token for test cases: %v", err)
	}

	// Generate an expired token for testing
	expiredExpiresIn := -1 * time.Minute // Expires 1 minute ago
	expiredToken, err := auth.MakeJWT(validUserID, jwtSecret, expiredExpiresIn)
	if err != nil {
		t.Fatalf("Failed to make an expired token for test cases: %v", err)
	}

	// Generate a token with invalid claims (malformed subject)
	malformedClaimsToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(validExpiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "chirpy",
		Subject:   "not-a-uuid", // Invalid UUID string
	})
	malformedTokenString, err := malformedClaimsToken.SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("Failed to sign malformed claims token for test cases: %v", err)
	}
	// Get an example error from uuid.Parse for comparison
	_, expectedParseErr := uuid.Parse("not-a-uuid")

	// Define a slice of test cases
	tests := []struct {
		name           string
		tokenString    string
		jwtSecret      string
		expectedUserID uuid.UUID
		expectedErr    error // The specific error we expect, or nil for success
		expectErr      bool
		errSubstring   string // Substring to check in the error message if expected
	}{
		{
			name:           "Valid Token",
			tokenString:    validToken,
			jwtSecret:      jwtSecret,
			expectedUserID: validUserID,
			expectedErr:    nil,
			expectErr:      false,
		},
		{
			name:           "Invalid Secret",
			tokenString:    validToken,
			jwtSecret:      "wrong-secret",
			expectedUserID: uuid.Nil, // Expecting nil UUID on error
			expectedErr:    nil,      // Expecting an error, but not checking a specific type
			expectErr:      true,
			errSubstring:   "signature", // Checking for substring related to signature issues
		},
		{
			name:           "Expired Token",
			tokenString:    expiredToken,
			jwtSecret:      jwtSecret,
			expectedUserID: uuid.Nil, // Expecting nil UUID on error
			expectedErr:    nil,      // Expecting an error, but not checking a specific type
			expectErr:      true,
			errSubstring:   "expired", // Checking for substring related to expiration
		},
		{
			name:           "Token with Malformed Subject",
			tokenString:    malformedTokenString,
			jwtSecret:      jwtSecret,
			expectedUserID: uuid.Nil,         // Expecting nil UUID on error
			expectedErr:    expectedParseErr, // Expecting the error returned by uuid.Parse
			expectErr:      true,
		},
		{
			name:           "Empty Token String",
			tokenString:    "",
			jwtSecret:      jwtSecret,
			expectedUserID: uuid.Nil, // Expecting nil UUID on error
			expectedErr:    nil,      // Expecting an error, but not checking a specific type
			expectErr:      true,
			errSubstring:   "segments", // Checking for substring related to invalid segments
		},
	}

	// Iterate over the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validatedUserID, err := auth.ValidateJWT(tt.tokenString, tt.jwtSecret)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				} else {
					// If a specific error instance is expected, check with errors.Is
					if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
						t.Errorf("Expected error to contain %v, but got %v", tt.expectedErr, err)
					}
					// If checking for a substring, check that
					if tt.errSubstring != "" && !strings.Contains(err.Error(), tt.errSubstring) {
						t.Errorf("Expected error message to contain '%s', but got '%v'", tt.errSubstring, err)
					}
				}
			} else { // Expecting no error
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
				if validatedUserID != tt.expectedUserID {
					t.Errorf("ValidateJWT returned incorrect UserID. Expected %s, got %s", tt.expectedUserID, validatedUserID)
				}
			}
		})
	}
}

// TestGetBearerToken tests the GetBearerToken function using test cases.
func TestGetBearerToken(t *testing.T) {
	// Define a slice of test cases
	tests := []struct {
		name          string
		headers       http.Header
		expectedToken string
		expectErr     bool
	}{
		{ // Added the missing opening brace
			name:          "No Authorization header",
			headers:       http.Header{},
			expectedToken: "", // Expected token should be empty on error
			expectErr:     true,
		}, // Added the missing closing brace
		{
			name: "Malformed Authorization header (no Bearer)",
			headers: http.Header{
				"Authorization": []string{"Basic abcdef"},
			},
			expectedToken: "", // Expected token should be empty on error
			expectErr:     true,
		},
		{
			name: "Valid Bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer my-token"},
			},
			expectedToken: "my-token",
			expectErr:     false,
		},
	}

	// Iterate over the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { // t.Run allows for subtests with names
			token, err := auth.GetBearerToken(tt.headers)

			if tt.expectErr {
				// If we expect an error
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
				// When an error is expected, the returned token should typically be an empty string.
				if token != tt.expectedToken {
					t.Errorf("Expected token %s on error, but got %s", tt.expectedToken, token)
				}
				// Optional: You could also check the specific error message or type here
				// if you want to be more precise about which error you expect.
			} else {
				// If we do NOT expect an error
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
				// Only check the token value when no error is expected
				if token != tt.expectedToken {
					t.Errorf("Expected token %s, but got %s", tt.expectedToken, token)
				}
			}
		})
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
