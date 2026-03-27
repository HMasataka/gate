package config

import (
	"strings"
	"testing"
	"time"
)

// TestValidate_Default verifies that ForTest() produces a config that passes Validate.
func TestValidate_Default(t *testing.T) {
	cfg := ForTest()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no validation errors, got: %v", err)
	}
}

// TestValidate_InvalidPort tests that out-of-range port values are rejected.
func TestValidate_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{name: "zero", port: 0},
		{name: "too large", port: 70000},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ForTest()
			cfg.Server.Port = tc.port
			err := cfg.Validate()
			if err == nil {
				t.Fatalf("expected validation error for port=%d, got nil", tc.port)
			}
			if !strings.Contains(err.Error(), "Server.Port") {
				t.Errorf("expected error to mention Server.Port, got: %v", err)
			}
		})
	}
}

// TestValidate_PasswordLengthRange tests password length validation rules.
func TestValidate_PasswordLengthRange(t *testing.T) {
	tests := []struct {
		name        string
		minLen      int
		maxLen      int
		wantErrFrag string
	}{
		{
			name:        "min greater than max",
			minLen:      64,
			maxLen:      32,
			wantErrFrag: "PasswordMinLength",
		},
		{
			name:        "min below 8",
			minLen:      4,
			maxLen:      128,
			wantErrFrag: "PasswordMinLength",
		},
		{
			name:        "max above 128",
			minLen:      8,
			maxLen:      200,
			wantErrFrag: "PasswordMaxLength",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ForTest()
			cfg.Auth.PasswordMinLength = tc.minLen
			cfg.Auth.PasswordMaxLength = tc.maxLen
			err := cfg.Validate()
			if err == nil {
				t.Fatalf("expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErrFrag) {
				t.Errorf("expected error to contain %q, got: %v", tc.wantErrFrag, err)
			}
		})
	}
}

// TestValidate_TokenExpiry tests that out-of-range AccessTokenExpiry is rejected.
func TestValidate_TokenExpiry(t *testing.T) {
	tests := []struct {
		name   string
		expiry time.Duration
	}{
		{name: "too short (below 1m)", expiry: 30 * time.Second},
		{name: "too long (above 1h)", expiry: 2 * time.Hour},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ForTest()
			cfg.Token.AccessTokenExpiry = tc.expiry
			err := cfg.Validate()
			if err == nil {
				t.Fatalf("expected validation error for AccessTokenExpiry=%s, got nil", tc.expiry)
			}
			if !strings.Contains(err.Error(), "Token.AccessTokenExpiry") {
				t.Errorf("expected error to mention Token.AccessTokenExpiry, got: %v", err)
			}
		})
	}
}

// TestValidate_DatabaseConns tests that MaxIdleConns > MaxOpenConns is rejected.
func TestValidate_DatabaseConns(t *testing.T) {
	cfg := ForTest()
	cfg.Database.MaxOpenConns = 5
	cfg.Database.MaxIdleConns = 10 // exceeds MaxOpenConns
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error when MaxIdleConns > MaxOpenConns, got nil")
	}
	if !strings.Contains(err.Error(), "Database.MaxIdleConns") {
		t.Errorf("expected error to mention Database.MaxIdleConns, got: %v", err)
	}
}

// TestValidate_JWTAlgorithm tests that an unsupported JWT algorithm is rejected.
func TestValidate_JWTAlgorithm(t *testing.T) {
	cfg := ForTest()
	cfg.JWT.Algorithm = "HS256"
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for invalid JWT algorithm, got nil")
	}
	if !strings.Contains(err.Error(), "JWT.Algorithm") {
		t.Errorf("expected error to mention JWT.Algorithm, got: %v", err)
	}
}

// TestValidate_LogLevel tests that an unsupported log level is rejected.
func TestValidate_LogLevel(t *testing.T) {
	cfg := ForTest()
	cfg.Log.Level = "TRACE"
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for invalid log level, got nil")
	}
	if !strings.Contains(err.Error(), "Log.Level") {
		t.Errorf("expected error to mention Log.Level, got: %v", err)
	}
}

// TestValidate_CORSWildcard tests that a wildcard "*" in AllowedOrigins is rejected.
func TestValidate_CORSWildcard(t *testing.T) {
	cfg := ForTest()
	cfg.CORS.AllowedOrigins = []string{"https://example.com", "*"}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for wildcard CORS origin, got nil")
	}
	if !strings.Contains(err.Error(), "CORS.AllowedOrigins") {
		t.Errorf("expected error to mention CORS.AllowedOrigins, got: %v", err)
	}
}

// TestValidate_MultipleErrors tests that multiple validation errors are returned together.
func TestValidate_MultipleErrors(t *testing.T) {
	cfg := ForTest()
	// Introduce multiple invalid values
	cfg.Server.Port = 0          // invalid: below 1
	cfg.JWT.Algorithm = "HS256"  // invalid: not ES256 or RS256
	cfg.Log.Level = "VERBOSE"    // invalid: not DEBUG/INFO/WARN/ERROR

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected multiple validation errors, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Server.Port") {
		t.Errorf("expected error to mention Server.Port, got: %v", errMsg)
	}
	if !strings.Contains(errMsg, "JWT.Algorithm") {
		t.Errorf("expected error to mention JWT.Algorithm, got: %v", errMsg)
	}
	if !strings.Contains(errMsg, "Log.Level") {
		t.Errorf("expected error to mention Log.Level, got: %v", errMsg)
	}
}
