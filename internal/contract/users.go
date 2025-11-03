package contract

import (
	"context"
	"net/http"
	"time"

	"github.com/rom8726/floxy-manager/internal/domain"
)

type UsersUseCase interface {
	Login(
		ctx context.Context,
		username, password string,
	) (accessToken, refreshToken, sessionID string, isTmpPassword bool, err error)
	LoginReissue(
		ctx context.Context,
		currRefreshToken string,
	) (accessToken, refreshToken string, err error)
	SSOInitiate(ctx context.Context, providerName string) (redirectURL string, err error)
	SSOCallback(
		ctx context.Context,
		providerName string, req *http.Request, response, state string,
	) (accessToken, refreshToken string, expiresIn int, err error)
	GetSSOProviders(ctx context.Context) ([]SSOProvider, error)
	GetSSOMetadata(ctx context.Context, providerName string) ([]byte, error)
	List(ctx context.Context) ([]domain.User, error)
	GetByID(ctx context.Context, id domain.UserID) (domain.User, error)
	Create(
		ctx context.Context,
		currentUser domain.User,
		username, email, password string,
		isSuperuser bool,
	) (domain.User, error)
	SetSuperuserStatus(ctx context.Context, id domain.UserID, isSuperuser bool) (domain.User, error)
	SetActiveStatus(ctx context.Context, id domain.UserID, isActive bool) (domain.User, error)
	Delete(ctx context.Context, id domain.UserID) error
	UpdatePassword(ctx context.Context, id domain.UserID, oldPassword, newPassword string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	Setup2FA(ctx context.Context, userID domain.UserID) (secret, qrURL, qrImage string, err error)
	Confirm2FA(ctx context.Context, userID domain.UserID, code string) error
	Send2FACode(ctx context.Context, userID domain.UserID, action string) error
	Disable2FA(ctx context.Context, userID domain.UserID, emailCode string) error
	Reset2FA(ctx context.Context, userID domain.UserID, emailCode string) (secret, qrURL, qrImage string, err error)
	Verify2FA(ctx context.Context, code, sessionID string) (accessToken, refreshToken string, expiresIn int, err error)
	VerifyTOTP(ctx context.Context, userID domain.UserID, code string) error
	InitiateTOTPApproval(ctx context.Context, userID domain.UserID) (sessionID string, err error)
	UpdateLicenseAcceptance(ctx context.Context, userID domain.UserID, accepted bool) error
	VerifyPassword(ctx context.Context, userID domain.UserID, password string) error
}

type UsersRepository interface {
	FetchByIDs(ctx context.Context, ids []domain.UserID) ([]domain.User, error)
	Create(ctx context.Context, user domain.UserDTO) (domain.User, error)
	GetByID(ctx context.Context, id domain.UserID) (domain.User, error)
	GetByUsername(ctx context.Context, username string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	ExistsByID(ctx context.Context, id domain.UserID) (bool, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id domain.UserID) error
	List(ctx context.Context) ([]domain.User, error)
	UpdateLastLogin(ctx context.Context, id domain.UserID) error
	UpdatePassword(ctx context.Context, id domain.UserID, passwordHash string) error
	Update2FA(ctx context.Context, id domain.UserID, enabled bool, secret string, confirmedAt *time.Time) error
}
