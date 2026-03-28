package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/HMasataka/gate/internal/config"
	"github.com/HMasataka/gate/internal/domain"
)

type mockClientRepo struct {
	client *domain.OAuthClient
	err    error
}

func (m *mockClientRepo) Create(_ context.Context, _ *domain.OAuthClient) error { return nil }
func (m *mockClientRepo) GetByID(_ context.Context, _ string) (*domain.OAuthClient, error) {
	return m.client, m.err
}
func (m *mockClientRepo) Update(_ context.Context, _ *domain.OAuthClient) error { return nil }
func (m *mockClientRepo) Delete(_ context.Context, _ string) error              { return nil }
func (m *mockClientRepo) List(_ context.Context, _, _ int) ([]domain.OAuthClient, int, error) {
	return nil, 0, nil
}
func (m *mockClientRepo) ListByOwner(_ context.Context, _ string, _, _ int) ([]domain.OAuthClient, int, error) {
	return nil, 0, nil
}

type mockCodeRepo struct {
	code *domain.AuthorizationCode
	err  error
}

func (m *mockCodeRepo) Create(_ context.Context, _ *domain.AuthorizationCode) error { return nil }
func (m *mockCodeRepo) GetByCode(_ context.Context, _ string) (*domain.AuthorizationCode, error) {
	return m.code, m.err
}
func (m *mockCodeRepo) MarkUsed(_ context.Context, _ string) error        { return nil }
func (m *mockCodeRepo) DeleteExpired(_ context.Context) (int64, error)    { return 0, nil }

type mockRandom struct{}

func (m *mockRandom) GenerateToken(_ int) (string, error) { return "test-token", nil }
func (m *mockRandom) GenerateUUID() string                { return "test-uuid" }

type mockTxRunner struct{}

func (m *mockTxRunner) RunInTx(_ context.Context, fn func(context.Context) error) error {
	return fn(context.Background())
}

func newTestOAuthUsecase(clientRepo domain.OAuthClientRepository, codeRepo domain.AuthorizationCodeRepository) *OAuthUsecase {
	return NewOAuthUsecase(
		clientRepo,
		codeRepo,
		nil,
		&mockRandom{},
		config.OAuthConfig{},
		config.TokenConfig{AuthCodeExpiry: 10 * time.Minute},
		&mockTxRunner{},
	)
}

func TestAuthorize_EmptyUserID_ReturnsErrSessionRequired(t *testing.T) {
	clientRepo := &mockClientRepo{
		client: &domain.OAuthClient{
			ID:           "client-1",
			RedirectURIs: []string{"http://localhost:3000/callback"},
		},
	}
	codeRepo := &mockCodeRepo{}
	uc := newTestOAuthUsecase(clientRepo, codeRepo)

	_, err := uc.Authorize(context.Background(), "", "client-1", "http://localhost:3000/callback", "code", "openid", "state", "", "")
	if !errors.Is(err, domain.ErrSessionRequired) {
		t.Errorf("expected ErrSessionRequired, got %v", err)
	}
}

func TestAuthorize_ValidUserID_ReturnsCode(t *testing.T) {
	clientRepo := &mockClientRepo{
		client: &domain.OAuthClient{
			ID:           "client-1",
			RedirectURIs: []string{"http://localhost:3000/callback"},
		},
	}
	codeRepo := &mockCodeRepo{}
	uc := newTestOAuthUsecase(clientRepo, codeRepo)

	code, err := uc.Authorize(context.Background(), "user-123", "client-1", "http://localhost:3000/callback", "code", "openid", "state", "", "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if code == "" {
		t.Error("expected non-empty code")
	}
}

func TestAuthorize_InvalidClient_ReturnsError(t *testing.T) {
	clientRepo := &mockClientRepo{err: errors.New("not found")}
	codeRepo := &mockCodeRepo{}
	uc := newTestOAuthUsecase(clientRepo, codeRepo)

	_, err := uc.Authorize(context.Background(), "user-123", "bad-client", "http://localhost:3000/callback", "code", "", "", "", "")
	if !errors.Is(err, domain.ErrInvalidClient) {
		t.Errorf("expected ErrInvalidClient, got %v", err)
	}
}
