package config

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	JWT      JWTConfig
	Token    TokenConfig
	OAuth    OAuthConfig
	MFA      MFAConfig
	Session  SessionConfig
	CORS     CORSConfig
	Argon2   Argon2Config
	Mailer   MailerConfig
	Log      LogConfig
	Metrics  MetricsConfig
}

type ServerConfig struct {
	Port              int           `env:"SERVER_PORT"               envDefault:"8080"`
	ReadTimeout       time.Duration `env:"SERVER_READ_TIMEOUT"       envDefault:"15s"`
	WriteTimeout      time.Duration `env:"SERVER_WRITE_TIMEOUT"      envDefault:"15s"`
	IdleTimeout       time.Duration `env:"SERVER_IDLE_TIMEOUT"       envDefault:"60s"`
	ReadHeaderTimeout time.Duration `env:"SERVER_READ_HEADER_TIMEOUT" envDefault:"5s"`
	ShutdownTimeout   time.Duration `env:"SHUTDOWN_TIMEOUT"          envDefault:"30s"`
}

type DatabaseConfig struct {
	URL             string        `env:"DATABASE_URL,required"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS"     envDefault:"25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS"     envDefault:"5"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME"  envDefault:"1h"`
	ConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"30m"`
	ConnectRetries  int           `env:"DB_CONNECT_RETRIES"    envDefault:"5"`
	MigrateTimeout  time.Duration `env:"DB_MIGRATE_TIMEOUT"    envDefault:"60s"`
}

type RedisConfig struct {
	URL          string        `env:"REDIS_URL,required"`
	PoolSize     int           `env:"REDIS_POOL_SIZE"      envDefault:"10"`
	MinIdleConns int           `env:"REDIS_MIN_IDLE_CONNS" envDefault:"5"`
	DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT"   envDefault:"5s"`
	ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT"   envDefault:"3s"`
	WriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT"  envDefault:"3s"`
}

type AuthConfig struct {
	PasswordMinLength       int           `env:"PASSWORD_MIN_LENGTH"       envDefault:"8"`
	PasswordMaxLength       int           `env:"PASSWORD_MAX_LENGTH"       envDefault:"128"`
	EmailVerificationExpiry time.Duration `env:"EMAIL_VERIFICATION_EXPIRY" envDefault:"24h"`
	PasswordResetExpiry     time.Duration `env:"PASSWORD_RESET_EXPIRY"     envDefault:"1h"`
	VerificationResendLimit int           `env:"VERIFICATION_RESEND_LIMIT" envDefault:"3"`
	LoginFailLockThreshold  int           `env:"LOGIN_FAIL_LOCK_THRESHOLD" envDefault:"5"`
	LoginFailLockDuration   time.Duration `env:"LOGIN_FAIL_LOCK_DURATION"  envDefault:"30m"`
	AccountPurgeDays        int           `env:"ACCOUNT_PURGE_DAYS"        envDefault:"30"`
}

type JWTConfig struct {
	PrivateKeyPath string `env:"JWT_PRIVATE_KEY_PATH" envDefault:""`
	PublicKeyPath  string `env:"JWT_PUBLIC_KEY_PATH"  envDefault:""`
	Algorithm      string `env:"JWT_ALGORITHM"        envDefault:"ES256"`
}

type TokenConfig struct {
	AccessTokenExpiry       time.Duration `env:"ACCESS_TOKEN_EXPIRY"         envDefault:"15m"`
	RefreshTokenExpiry      time.Duration `env:"REFRESH_TOKEN_EXPIRY"        envDefault:"168h"`
	IDTokenExpiry           time.Duration `env:"ID_TOKEN_EXPIRY"             envDefault:"1h"`
	AuthCodeExpiry          time.Duration `env:"AUTH_CODE_EXPIRY"            envDefault:"10m"`
	GracePeriod             time.Duration `env:"TOKEN_GRACE_PERIOD"          envDefault:"10s"`
	MaxCustomClaims         int           `env:"MAX_CUSTOM_CLAIMS"           envDefault:"20"`
	MaxCustomClaimValueSize int           `env:"MAX_CUSTOM_CLAIM_VALUE_SIZE"  envDefault:"1024"`
}

type OAuthConfig struct {
	MaxRedirectURIs int `env:"OAUTH_MAX_REDIRECT_URIS" envDefault:"10"`
}

type MFAConfig struct {
	RecoveryCodeCount  int `env:"MFA_RECOVERY_CODE_COUNT"  envDefault:"8"`
	RecoveryCodeLength int `env:"MFA_RECOVERY_CODE_LENGTH" envDefault:"8"`
	TOTPSkew           int `env:"MFA_TOTP_SKEW"            envDefault:"1"`
}

type SessionConfig struct {
	MaxSessions int `env:"MAX_SESSIONS" envDefault:"5"`
}

type CORSConfig struct {
	AllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:","`
	MaxAge         int      `env:"CORS_MAX_AGE"         envDefault:"300"`
}

type Argon2Config struct {
	Time        uint32 `env:"ARGON2_TIME"        envDefault:"1"`
	Memory      uint32 `env:"ARGON2_MEMORY"      envDefault:"65536"`
	Parallelism uint8  `env:"ARGON2_PARALLELISM" envDefault:"4"`
	KeyLength   uint32 `env:"ARGON2_KEY_LENGTH"  envDefault:"32"`
	SaltLength  uint32 `env:"ARGON2_SALT_LENGTH" envDefault:"16"`
	Semaphore   int    `env:"ARGON2_SEMAPHORE"   envDefault:"16"`
}

type MailerConfig struct {
	Type string `env:"MAILER_TYPE" envDefault:"stdout"`
}

type LogConfig struct {
	Level  string `env:"LOG_LEVEL"  envDefault:"INFO"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

type MetricsConfig struct {
	Enabled bool `env:"METRICS_ENABLED" envDefault:"true"`
}

// Load parses configuration from environment variables.
func Load() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// Validate checks all configuration values and returns all validation errors
// aggregated into a single error.
func (c *Config) Validate() error {
	var errs []error

	// Server.Port: 1-65535
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, fmt.Errorf("Server.Port must be between 1 and 65535, got %d", c.Server.Port))
	}

	// Server timeouts must be > 0
	if c.Server.ReadTimeout <= 0 {
		errs = append(errs, fmt.Errorf("Server.ReadTimeout must be > 0, got %s", c.Server.ReadTimeout))
	}
	if c.Server.WriteTimeout <= 0 {
		errs = append(errs, fmt.Errorf("Server.WriteTimeout must be > 0, got %s", c.Server.WriteTimeout))
	}
	if c.Server.IdleTimeout <= 0 {
		errs = append(errs, fmt.Errorf("Server.IdleTimeout must be > 0, got %s", c.Server.IdleTimeout))
	}
	if c.Server.ReadHeaderTimeout <= 0 {
		errs = append(errs, fmt.Errorf("Server.ReadHeaderTimeout must be > 0, got %s", c.Server.ReadHeaderTimeout))
	}
	if c.Server.ShutdownTimeout <= 0 {
		errs = append(errs, fmt.Errorf("Server.ShutdownTimeout must be > 0, got %s", c.Server.ShutdownTimeout))
	}

	// Auth.PasswordMinLength: 8-128
	if c.Auth.PasswordMinLength < 8 || c.Auth.PasswordMinLength > 128 {
		errs = append(errs, fmt.Errorf("Auth.PasswordMinLength must be between 8 and 128, got %d", c.Auth.PasswordMinLength))
	}

	// Auth.PasswordMaxLength: 8-128
	if c.Auth.PasswordMaxLength < 8 || c.Auth.PasswordMaxLength > 128 {
		errs = append(errs, fmt.Errorf("Auth.PasswordMaxLength must be between 8 and 128, got %d", c.Auth.PasswordMaxLength))
	}

	// PasswordMinLength <= PasswordMaxLength (only check when both are individually valid)
	if c.Auth.PasswordMinLength >= 8 && c.Auth.PasswordMinLength <= 128 &&
		c.Auth.PasswordMaxLength >= 8 && c.Auth.PasswordMaxLength <= 128 &&
		c.Auth.PasswordMinLength > c.Auth.PasswordMaxLength {
		errs = append(errs, fmt.Errorf("Auth.PasswordMinLength (%d) must be <= Auth.PasswordMaxLength (%d)", c.Auth.PasswordMinLength, c.Auth.PasswordMaxLength))
	}

	// Token.AccessTokenExpiry: 1m-1h
	if c.Token.AccessTokenExpiry < time.Minute || c.Token.AccessTokenExpiry > time.Hour {
		errs = append(errs, fmt.Errorf("Token.AccessTokenExpiry must be between 1m and 1h, got %s", c.Token.AccessTokenExpiry))
	}

	// Token.RefreshTokenExpiry: 1h-2160h (90 days)
	if c.Token.RefreshTokenExpiry < time.Hour || c.Token.RefreshTokenExpiry > 2160*time.Hour {
		errs = append(errs, fmt.Errorf("Token.RefreshTokenExpiry must be between 1h and 2160h, got %s", c.Token.RefreshTokenExpiry))
	}

	// Token.IDTokenExpiry: 5m-24h
	if c.Token.IDTokenExpiry < 5*time.Minute || c.Token.IDTokenExpiry > 24*time.Hour {
		errs = append(errs, fmt.Errorf("Token.IDTokenExpiry must be between 5m and 24h, got %s", c.Token.IDTokenExpiry))
	}

	// Token.AuthCodeExpiry: 1m-30m
	if c.Token.AuthCodeExpiry < time.Minute || c.Token.AuthCodeExpiry > 30*time.Minute {
		errs = append(errs, fmt.Errorf("Token.AuthCodeExpiry must be between 1m and 30m, got %s", c.Token.AuthCodeExpiry))
	}

	// Token.GracePeriod: 0s-60s
	if c.Token.GracePeriod < 0 || c.Token.GracePeriod > 60*time.Second {
		errs = append(errs, fmt.Errorf("Token.GracePeriod must be between 0s and 60s, got %s", c.Token.GracePeriod))
	}

	// Database.MaxOpenConns: 1-1000
	if c.Database.MaxOpenConns < 1 || c.Database.MaxOpenConns > 1000 {
		errs = append(errs, fmt.Errorf("Database.MaxOpenConns must be between 1 and 1000, got %d", c.Database.MaxOpenConns))
	}

	// Database.MaxIdleConns: 0-MaxOpenConns
	if c.Database.MaxIdleConns < 0 || c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		errs = append(errs, fmt.Errorf("Database.MaxIdleConns must be between 0 and MaxOpenConns (%d), got %d", c.Database.MaxOpenConns, c.Database.MaxIdleConns))
	}

	// Database.ConnectRetries: 0-20
	if c.Database.ConnectRetries < 0 || c.Database.ConnectRetries > 20 {
		errs = append(errs, fmt.Errorf("Database.ConnectRetries must be between 0 and 20, got %d", c.Database.ConnectRetries))
	}

	// JWT.Algorithm: "ES256" or "RS256"
	if c.JWT.Algorithm != "ES256" && c.JWT.Algorithm != "RS256" {
		errs = append(errs, fmt.Errorf("JWT.Algorithm must be ES256 or RS256, got %q", c.JWT.Algorithm))
	}

	// Log.Level: "DEBUG", "INFO", "WARN", "ERROR"
	switch c.Log.Level {
	case "DEBUG", "INFO", "WARN", "ERROR":
		// valid
	default:
		errs = append(errs, fmt.Errorf("Log.Level must be one of DEBUG, INFO, WARN, ERROR, got %q", c.Log.Level))
	}

	// CORS.AllowedOrigins must not contain "*"
	if slices.Contains(c.CORS.AllowedOrigins, "*") {
		errs = append(errs, errors.New("CORS.AllowedOrigins must not contain wildcard \"*\""))
	}

	return errors.Join(errs...)
}

// ForTest returns a Config suitable for use in tests.
func ForTest() *Config {
	return &Config{
		Server: ServerConfig{
			Port:              8080,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			ShutdownTimeout:   30 * time.Second,
		},
		Database: DatabaseConfig{
			URL:             "postgres://test:test@localhost:5432/gate_test?sslmode=disable",
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: 30 * time.Minute,
			ConnectRetries:  3,
			MigrateTimeout:  30 * time.Second,
		},
		Redis: RedisConfig{
			URL:          "redis://localhost:6379/1",
			PoolSize:     5,
			MinIdleConns: 1,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
		Auth: AuthConfig{
			PasswordMinLength:       8,
			PasswordMaxLength:       128,
			EmailVerificationExpiry: 24 * time.Hour,
			PasswordResetExpiry:     time.Hour,
			VerificationResendLimit: 3,
			LoginFailLockThreshold:  5,
			LoginFailLockDuration:   30 * time.Minute,
			AccountPurgeDays:        30,
		},
		JWT: JWTConfig{
			Algorithm: "ES256",
		},
		Token: TokenConfig{
			AccessTokenExpiry:       15 * time.Minute,
			RefreshTokenExpiry:      168 * time.Hour,
			IDTokenExpiry:           time.Hour,
			AuthCodeExpiry:          10 * time.Minute,
			GracePeriod:             10 * time.Second,
			MaxCustomClaims:         20,
			MaxCustomClaimValueSize: 1024,
		},
		OAuth: OAuthConfig{
			MaxRedirectURIs: 10,
		},
		MFA: MFAConfig{
			RecoveryCodeCount:  8,
			RecoveryCodeLength: 8,
			TOTPSkew:           1,
		},
		Session: SessionConfig{
			MaxSessions: 5,
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
			MaxAge:         300,
		},
		Argon2: Argon2Config{
			Time:        1,
			Memory:      65536,
			Parallelism: 4,
			KeyLength:   32,
			SaltLength:  16,
			Semaphore:   16,
		},
		Mailer: MailerConfig{
			Type: "stdout",
		},
		Log: LogConfig{
			Level:  "DEBUG",
			Format: "text",
		},
		Metrics: MetricsConfig{
			Enabled: true,
		},
	}
}
