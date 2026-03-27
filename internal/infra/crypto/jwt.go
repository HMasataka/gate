package crypto

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"maps"
	"math/big"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/HMasataka/gate/internal/config"
)

type JWTManagerImpl struct {
	privateKey crypto.Signer
	publicKey  crypto.PublicKey
	kid        string
	algorithm  string
	method     jwt.SigningMethod
	tokenCfg   config.TokenConfig
}

func NewJWTManager(jwtCfg config.JWTConfig, tokenCfg config.TokenConfig) (*JWTManagerImpl, error) {
	m := &JWTManagerImpl{
		algorithm: jwtCfg.Algorithm,
		tokenCfg:  tokenCfg,
	}

	switch jwtCfg.Algorithm {
	case "ES256":
		m.method = jwt.SigningMethodES256
	case "RS256":
		m.method = jwt.SigningMethodRS256
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", jwtCfg.Algorithm)
	}

	if jwtCfg.PrivateKeyPath != "" {
		if err := m.loadKeys(jwtCfg); err != nil {
			return nil, err
		}
	} else {
		if err := m.generateDevKeys(); err != nil {
			return nil, err
		}
	}

	m.kid = computeKID(m.publicKey)

	return m, nil
}

func (m *JWTManagerImpl) loadKeys(cfg config.JWTConfig) error {
	privPEM, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("read private key: %w", err)
	}

	block, _ := pem.Decode(privPEM)
	if block == nil {
		return fmt.Errorf("decode private key PEM: no valid block found")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Fallback to algorithm-specific parsers.
		switch cfg.Algorithm {
		case "ES256":
			privKey, err = x509.ParseECPrivateKey(block.Bytes)
		case "RS256":
			privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		}
		if err != nil {
			return fmt.Errorf("parse private key: %w", err)
		}
	}

	signer, ok := privKey.(crypto.Signer)
	if !ok {
		return fmt.Errorf("private key does not implement crypto.Signer")
	}

	m.privateKey = signer

	if cfg.PublicKeyPath != "" {
		pubPEM, err := os.ReadFile(cfg.PublicKeyPath)
		if err != nil {
			return fmt.Errorf("read public key: %w", err)
		}

		pubBlock, _ := pem.Decode(pubPEM)
		if pubBlock == nil {
			return fmt.Errorf("decode public key PEM: no valid block found")
		}

		pubKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
		if err != nil {
			return fmt.Errorf("parse public key: %w", err)
		}

		m.publicKey = pubKey
	} else {
		m.publicKey = signer.Public()
	}

	return nil
}

func (m *JWTManagerImpl) generateDevKeys() error {
	switch m.algorithm {
	case "ES256":
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return fmt.Errorf("generate EC dev key: %w", err)
		}
		m.privateKey = key
		m.publicKey = &key.PublicKey
	case "RS256":
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fmt.Errorf("generate RSA dev key: %w", err)
		}
		m.privateKey = key
		m.publicKey = &key.PublicKey
	default:
		return fmt.Errorf("unsupported algorithm for dev keys: %s", m.algorithm)
	}

	return nil
}

func computeKID(pub crypto.PublicKey) string {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "unknown"
	}

	hash := sha256.Sum256(der)

	return hex.EncodeToString(hash[:8])
}

func (m *JWTManagerImpl) GenerateAccessToken(_ context.Context, claims map[string]any) (string, error) {
	return m.generateToken(claims, m.tokenCfg.AccessTokenExpiry)
}

func (m *JWTManagerImpl) GenerateIDToken(_ context.Context, claims map[string]any) (string, error) {
	return m.generateToken(claims, m.tokenCfg.IDTokenExpiry)
}

func (m *JWTManagerImpl) generateToken(claims map[string]any, expiry time.Duration) (string, error) {
	now := time.Now()

	mapClaims := jwt.MapClaims{}
	maps.Copy(mapClaims, claims)

	mapClaims["iat"] = now.Unix()
	mapClaims["exp"] = now.Add(expiry).Unix()
	mapClaims["jti"] = uuid.New().String()

	token := jwt.NewWithClaims(m.method, mapClaims)
	token.Header["kid"] = m.kid

	return token.SignedString(m.privateKey)
}

func (m *JWTManagerImpl) ValidateToken(_ context.Context, tokenString string) (map[string]any, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != m.method.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}

		kid, ok := t.Header["kid"].(string)
		if !ok || kid != m.kid {
			return nil, fmt.Errorf("unknown kid: %v", t.Header["kid"])
		}

		return m.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return map[string]any(mapClaims), nil
}

func (m *JWTManagerImpl) JWKS() map[string]any {
	key := m.buildJWK()

	return map[string]any{
		"keys": []map[string]any{key},
	}
}

func (m *JWTManagerImpl) buildJWK() map[string]any {
	switch pub := m.publicKey.(type) {
	case *ecdsa.PublicKey:
		byteLen := (pub.Curve.Params().BitSize + 7) / 8

		xBytes := pub.X.Bytes()
		yBytes := pub.Y.Bytes()

		// Pad to fixed length.
		xPadded := make([]byte, byteLen)
		yPadded := make([]byte, byteLen)
		copy(xPadded[byteLen-len(xBytes):], xBytes)
		copy(yPadded[byteLen-len(yBytes):], yBytes)

		return map[string]any{
			"kty": "EC",
			"crv": "P-256",
			"x":   base64.RawURLEncoding.EncodeToString(xPadded),
			"y":   base64.RawURLEncoding.EncodeToString(yPadded),
			"kid": m.kid,
			"use": "sig",
			"alg": m.algorithm,
		}
	case *rsa.PublicKey:
		nBytes := pub.N.Bytes()
		eBytes := big.NewInt(int64(pub.E)).Bytes()

		return map[string]any{
			"kty": "RSA",
			"n":   base64.RawURLEncoding.EncodeToString(nBytes),
			"e":   base64.RawURLEncoding.EncodeToString(eBytes),
			"kid": m.kid,
			"use": "sig",
			"alg": m.algorithm,
		}
	default:
		return map[string]any{}
	}
}
