package captcha

import "context"

type NoOpCaptchaVerifier struct{}

func (v *NoOpCaptchaVerifier) Verify(ctx context.Context, token string) (bool, error) {
	return true, nil
}
