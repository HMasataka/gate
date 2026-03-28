package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/HMasataka/gate/internal/config"
	"github.com/HMasataka/gate/internal/domain"
)

type ClientUsecase struct {
	clientRepo domain.OAuthClientRepository
	random     domain.RandomGenerator
	oauthCfg   config.OAuthConfig
}

func NewClientUsecase(
	clientRepo domain.OAuthClientRepository,
	random domain.RandomGenerator,
	oauthCfg config.OAuthConfig,
) *ClientUsecase {
	return &ClientUsecase{
		clientRepo: clientRepo,
		random:     random,
		oauthCfg:   oauthCfg,
	}
}

func (u *ClientUsecase) Register(
	ctx context.Context,
	name, clientType string,
	redirectURIs, scopes, grantTypes []string,
) (*domain.OAuthClient, error) {
	if len(redirectURIs) > u.oauthCfg.MaxRedirectURIs {
		return nil, domain.ErrInvalidRedirectURI
	}

	for _, uri := range redirectURIs {
		if !isAllowedRedirectURI(uri) {
			return nil, domain.ErrInvalidRedirectURI
		}
	}

	clientID := u.random.GenerateUUID()

	rawSecret, err := u.random.GenerateToken(32)
	if err != nil {
		return nil, err
	}

	hashedSecret := sha256hex(rawSecret)

	now := time.Now()

	client := &domain.OAuthClient{
		ID:            clientID,
		Secret:        hashedSecret,
		Name:          name,
		Type:          domain.ClientType(clientType),
		RedirectURIs:  redirectURIs,
		AllowedScopes: scopes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := u.clientRepo.Create(ctx, client); err != nil {
		return nil, err
	}

	// Expose the plain secret only at registration time.
	client.Secret = rawSecret

	return client, nil
}

func (u *ClientUsecase) Get(ctx context.Context, id string) (*domain.OAuthClient, error) {
	return u.clientRepo.GetByID(ctx, id)
}

func (u *ClientUsecase) List(ctx context.Context, offset, limit int) ([]domain.OAuthClient, int, error) {
	return u.clientRepo.List(ctx, offset, limit)
}

func (u *ClientUsecase) Update(
	ctx context.Context,
	id, name string,
	redirectURIs, scopes []string,
) (*domain.OAuthClient, error) {
	client, err := u.clientRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(redirectURIs) > u.oauthCfg.MaxRedirectURIs {
		return nil, domain.ErrInvalidRedirectURI
	}

	for _, uri := range redirectURIs {
		if !isAllowedRedirectURI(uri) {
			return nil, domain.ErrInvalidRedirectURI
		}
	}

	client.Name = name
	client.RedirectURIs = redirectURIs
	client.AllowedScopes = scopes
	client.UpdatedAt = time.Now()

	if err := u.clientRepo.Update(ctx, client); err != nil {
		return nil, err
	}

	return client, nil
}

func (u *ClientUsecase) Delete(ctx context.Context, id string) error {
	return u.clientRepo.Delete(ctx, id)
}

func (u *ClientUsecase) RotateSecret(ctx context.Context, id string) (*domain.OAuthClient, error) {
	client, err := u.clientRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	rawSecret, err := u.random.GenerateToken(32)
	if err != nil {
		return nil, err
	}

	client.Secret = sha256hex(rawSecret)
	client.UpdatedAt = time.Now()

	if err := u.clientRepo.Update(ctx, client); err != nil {
		return nil, err
	}

	// Expose the plain secret only at rotation time.
	client.Secret = rawSecret

	return client, nil
}

func (u *ClientUsecase) RegisterForUser(
	ctx context.Context,
	ownerID, name, clientType string,
	redirectURIs, scopes, grantTypes []string,
) (*domain.OAuthClient, error) {
	if len(redirectURIs) > u.oauthCfg.MaxRedirectURIs {
		return nil, domain.ErrInvalidRedirectURI
	}

	for _, uri := range redirectURIs {
		if !isAllowedRedirectURI(uri) {
			return nil, domain.ErrInvalidRedirectURI
		}
	}

	clientID := u.random.GenerateUUID()

	rawSecret, err := u.random.GenerateToken(32)
	if err != nil {
		return nil, err
	}

	hashedSecret := sha256hex(rawSecret)

	now := time.Now()

	client := &domain.OAuthClient{
		ID:            clientID,
		Secret:        hashedSecret,
		Name:          name,
		Type:          domain.ClientType(clientType),
		OwnerID:       ownerID,
		RedirectURIs:  redirectURIs,
		AllowedScopes: scopes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := u.clientRepo.Create(ctx, client); err != nil {
		return nil, err
	}

	client.Secret = rawSecret

	return client, nil
}

func (u *ClientUsecase) ListByOwner(ctx context.Context, ownerID string, offset, limit int) ([]domain.OAuthClient, int, error) {
	return u.clientRepo.ListByOwner(ctx, ownerID, offset, limit)
}

func (u *ClientUsecase) GetOwned(ctx context.Context, ownerID, clientID string) (*domain.OAuthClient, error) {
	client, err := u.clientRepo.GetByID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if client.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	return client, nil
}

func (u *ClientUsecase) UpdateOwned(
	ctx context.Context,
	ownerID, clientID, name string,
	redirectURIs, scopes []string,
) (*domain.OAuthClient, error) {
	client, err := u.clientRepo.GetByID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if client.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	if len(redirectURIs) > u.oauthCfg.MaxRedirectURIs {
		return nil, domain.ErrInvalidRedirectURI
	}

	for _, uri := range redirectURIs {
		if !isAllowedRedirectURI(uri) {
			return nil, domain.ErrInvalidRedirectURI
		}
	}

	client.Name = name
	client.RedirectURIs = redirectURIs
	client.AllowedScopes = scopes
	client.UpdatedAt = time.Now()

	if err := u.clientRepo.Update(ctx, client); err != nil {
		return nil, err
	}

	return client, nil
}

func (u *ClientUsecase) DeleteOwned(ctx context.Context, ownerID, clientID string) error {
	client, err := u.clientRepo.GetByID(ctx, clientID)
	if err != nil {
		return err
	}

	if client.OwnerID != ownerID {
		return domain.ErrForbidden
	}

	return u.clientRepo.Delete(ctx, clientID)
}

func (u *ClientUsecase) RotateSecretOwned(ctx context.Context, ownerID, clientID string) (*domain.OAuthClient, error) {
	client, err := u.clientRepo.GetByID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if client.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	rawSecret, err := u.random.GenerateToken(32)
	if err != nil {
		return nil, err
	}

	client.Secret = sha256hex(rawSecret)
	client.UpdatedAt = time.Now()

	if err := u.clientRepo.Update(ctx, client); err != nil {
		return nil, err
	}

	client.Secret = rawSecret

	return client, nil
}

// isAllowedRedirectURI returns true for https:// URIs and http://localhost URIs.
func isAllowedRedirectURI(uri string) bool {
	if strings.HasPrefix(uri, "https://") {
		return true
	}

	if strings.HasPrefix(uri, "http://localhost") {
		return true
	}

	return false
}
