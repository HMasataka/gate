package social

import (
	"context"

	"github.com/HMasataka/gate/internal/usecase"
)

// ProviderAdapter wraps a SocialProvider and adapts it to the usecase.SocialProvider interface.
type ProviderAdapter struct {
	inner SocialProvider
}

// NewProviderAdapter wraps an infra SocialProvider for use in the usecase layer.
func NewProviderAdapter(p SocialProvider) *ProviderAdapter {
	return &ProviderAdapter{inner: p}
}

// Name delegates to the inner provider.
func (a *ProviderAdapter) Name() string {
	return a.inner.Name()
}

// AuthURL delegates to the inner provider.
func (a *ProviderAdapter) AuthURL(state string) string {
	return a.inner.AuthURL(state)
}

// ExchangeCode delegates to the inner provider and maps UserInfo to usecase.SocialUserInfo.
func (a *ProviderAdapter) ExchangeCode(ctx context.Context, code string) (*usecase.SocialUserInfo, error) {
	info, err := a.inner.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return &usecase.SocialUserInfo{
		ProviderUserID: info.ProviderUserID,
		Email:          info.Email,
		Name:           info.Name,
		AvatarURL:      info.AvatarURL,
		EmailVerified:  info.EmailVerified,
	}, nil
}
