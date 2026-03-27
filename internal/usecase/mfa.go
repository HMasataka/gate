package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/HMasataka/gate/internal/config"
	"github.com/HMasataka/gate/internal/domain"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type MFAUsecase struct {
	userRepo domain.UserRepository
	random   domain.RandomGenerator
	mfaCfg   config.MFAConfig
}

func NewMFAUsecase(userRepo domain.UserRepository, random domain.RandomGenerator, mfaCfg config.MFAConfig) *MFAUsecase {
	return &MFAUsecase{
		userRepo: userRepo,
		random:   random,
		mfaCfg:   mfaCfg,
	}
}

// SetupTOTP generates a new TOTP secret for the user and returns the provisioning URI
// and base32 secret. Does NOT enable TOTP yet; stores pending secret with TOTPEnabled=false.
// Returns ErrMFAAlreadyEnabled if TOTP is already enabled.
func (u *MFAUsecase) SetupTOTP(ctx context.Context, userID string) (provisioningURI, secret string, err error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	if user.TOTPEnabled {
		return "", "", domain.ErrMFAAlreadyEnabled
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Gate",
		AccountName: user.Email,
		SecretSize:  20,
	})
	if err != nil {
		return "", "", err
	}

	user.TOTPSecret = key.Secret()
	user.TOTPEnabled = false
	user.UpdatedAt = time.Now()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return "", "", err
	}

	return key.URL(), key.Secret(), nil
}

// ConfirmTOTP verifies the user's first TOTP code and enables MFA.
// Generates RecoveryCodeCount recovery codes and returns them for the user to save.
// Returns ErrInvalidMFACode if the code is wrong.
func (u *MFAUsecase) ConfirmTOTP(ctx context.Context, userID, code string) (recoveryCodes []string, err error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	valid, err := totp.ValidateCustom(code, user.TOTPSecret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      uint(u.mfaCfg.TOTPSkew),
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, domain.ErrInvalidMFACode
	}

	if !valid {
		return nil, domain.ErrInvalidMFACode
	}

	codes := make([]string, u.mfaCfg.RecoveryCodeCount)
	for i := range codes {
		c, err := u.random.GenerateToken(u.mfaCfg.RecoveryCodeLength)
		if err != nil {
			return nil, err
		}
		codes[i] = c
	}

	user.TOTPEnabled = true
	user.RecoveryCodes = codes
	user.UpdatedAt = time.Now()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return codes, nil
}

// VerifyTOTP validates a TOTP code during login.
// Returns ErrInvalidMFACode if wrong.
func (u *MFAUsecase) VerifyTOTP(ctx context.Context, userID, code string) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	valid, err := totp.ValidateCustom(code, user.TOTPSecret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      uint(u.mfaCfg.TOTPSkew),
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return domain.ErrInvalidMFACode
	}

	if !valid {
		return domain.ErrInvalidMFACode
	}

	return nil
}

// DisableTOTP disables TOTP for the user after verifying their password.
// Clears TOTPSecret, sets TOTPEnabled=false, and clears RecoveryCodes.
func (u *MFAUsecase) DisableTOTP(ctx context.Context, userID, password string, hasher domain.PasswordHasher) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	match, err := hasher.Compare(ctx, password, user.PasswordHash)
	if err != nil {
		return err
	}

	if !match {
		return domain.ErrInvalidCredentials
	}

	user.TOTPSecret = ""
	user.TOTPEnabled = false
	user.RecoveryCodes = nil
	user.UpdatedAt = time.Now()

	return u.userRepo.Update(ctx, user)
}

// RegenerateRecoveryCodes generates a new set of recovery codes, replacing old ones.
// Requires TOTP to be enabled. Returns ErrMFANotEnabled if TOTP is not enabled.
func (u *MFAUsecase) RegenerateRecoveryCodes(ctx context.Context, userID string) (recoveryCodes []string, err error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.TOTPEnabled {
		return nil, domain.ErrMFANotEnabled
	}

	codes := make([]string, u.mfaCfg.RecoveryCodeCount)
	for i := range codes {
		c, err := u.random.GenerateToken(u.mfaCfg.RecoveryCodeLength)
		if err != nil {
			return nil, err
		}
		codes[i] = c
	}

	user.RecoveryCodes = codes
	user.UpdatedAt = time.Now()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return codes, nil
}

// VerifyRecoveryCode validates a recovery code using case-insensitive comparison.
// Removes the used code from the list (one-time use).
// Returns ErrInvalidRecoveryCode if not found.
func (u *MFAUsecase) VerifyRecoveryCode(ctx context.Context, userID, code string) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	matchIdx := -1
	for i, rc := range user.RecoveryCodes {
		if strings.EqualFold(rc, code) {
			matchIdx = i
			break
		}
	}

	if matchIdx == -1 {
		return domain.ErrInvalidRecoveryCode
	}

	user.RecoveryCodes = append(user.RecoveryCodes[:matchIdx], user.RecoveryCodes[matchIdx+1:]...)
	user.UpdatedAt = time.Now()

	return u.userRepo.Update(ctx, user)
}
