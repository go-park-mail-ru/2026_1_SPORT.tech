package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/domain"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,30}$`)

type Service struct {
	accountRepository AccountRepository
	sessionRepository SessionRepository
	passwordHasher    PasswordHasher
	tokenGenerator    TokenGenerator
	clock             Clock
	sessionTTL        time.Duration
}

func NewService(
	accountRepository AccountRepository,
	sessionRepository SessionRepository,
	passwordHasher PasswordHasher,
	tokenGenerator TokenGenerator,
	clock Clock,
	sessionTTL time.Duration,
) *Service {
	return &Service{
		accountRepository: accountRepository,
		sessionRepository: sessionRepository,
		passwordHasher:    passwordHasher,
		tokenGenerator:    tokenGenerator,
		clock:             clock,
		sessionTTL:        sessionTTL,
	}
}

func (service *Service) Register(ctx context.Context, command RegisterCommand) (AuthResult, error) {
	command.Email = normalizeEmail(command.Email)
	command.Username = normalizeUsername(command.Username)
	if command.Role == "" {
		command.Role = domain.RoleClient
	}

	if err := validateRegisterCommand(command); err != nil {
		return AuthResult{}, err
	}

	passwordHash, err := service.passwordHasher.Hash(ctx, command.Password)
	if err != nil {
		return AuthResult{}, err
	}

	now := service.clock.Now()
	account, err := service.accountRepository.Create(ctx, CreateAccountParams{
		Email:        command.Email,
		Username:     command.Username,
		PasswordHash: passwordHash,
		Role:         command.Role,
		Status:       domain.StatusActive,
		Now:          now,
	})
	if err != nil {
		return AuthResult{}, err
	}

	return service.issueSession(ctx, account, now)
}

func (service *Service) Login(ctx context.Context, command LoginCommand) (AuthResult, error) {
	command.Email = normalizeEmail(command.Email)
	if err := validateLoginCommand(command); err != nil {
		return AuthResult{}, err
	}

	account, err := service.accountRepository.GetByEmail(ctx, command.Email)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return AuthResult{}, domain.ErrInvalidCredentials
		}

		return AuthResult{}, err
	}

	if err := account.CanAuthenticate(); err != nil {
		return AuthResult{}, err
	}

	if err := service.passwordHasher.Compare(ctx, account.PasswordHash, command.Password); err != nil {
		return AuthResult{}, domain.ErrInvalidCredentials
	}

	return service.issueSession(ctx, account, service.clock.Now())
}

func (service *Service) Logout(ctx context.Context, command LogoutCommand) error {
	if err := validateSessionToken(command.SessionToken); err != nil {
		return err
	}

	return service.sessionRepository.RevokeByHash(ctx, hashSessionToken(command.SessionToken))
}

func (service *Service) GetSession(ctx context.Context, query GetSessionQuery) (SessionResult, error) {
	if err := validateSessionToken(query.SessionToken); err != nil {
		return SessionResult{}, err
	}

	session, err := service.sessionRepository.GetByHash(ctx, hashSessionToken(query.SessionToken))
	if err != nil {
		return SessionResult{}, err
	}

	now := service.clock.Now()
	if !session.IsActive(now) {
		return SessionResult{}, domain.ErrSessionExpired
	}

	account, err := service.accountRepository.GetByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return SessionResult{}, domain.ErrInvalidCredentials
		}

		return SessionResult{}, err
	}

	if err := account.CanAuthenticate(); err != nil {
		return SessionResult{}, err
	}

	return SessionResult{
		Account: account,
		Session: session,
	}, nil
}

func (service *Service) issueSession(ctx context.Context, account domain.Account, now time.Time) (AuthResult, error) {
	sessionToken, err := service.tokenGenerator.NewToken(ctx)
	if err != nil {
		return AuthResult{}, err
	}

	session := domain.Session{
		IDHash:    hashSessionToken(sessionToken),
		UserID:    account.ID,
		ExpiresAt: now.Add(service.sessionTTL),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := service.sessionRepository.Create(ctx, session); err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		Account:          account,
		SessionToken:     sessionToken,
		SessionExpiresAt: session.ExpiresAt,
	}, nil
}

func validateRegisterCommand(command RegisterCommand) error {
	if !isValidEmail(command.Email) {
		return ErrInvalidEmail
	}

	if !usernamePattern.MatchString(command.Username) {
		return ErrInvalidUsername
	}

	if len(command.Password) < 8 {
		return ErrWeakPassword
	}

	if !command.Role.IsValid() {
		return domain.ErrInvalidRole
	}

	return nil
}

func validateLoginCommand(command LoginCommand) error {
	if !isValidEmail(command.Email) {
		return ErrInvalidEmail
	}

	if len(command.Password) < 8 {
		return ErrWeakPassword
	}

	return nil
}

func validateSessionToken(sessionToken string) error {
	if strings.TrimSpace(sessionToken) == "" {
		return ErrMissingSessionToken
	}

	return nil
}

func isValidEmail(email string) bool {
	if len(email) > 254 {
		return false
	}

	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeUsername(username string) string {
	return strings.TrimSpace(username)
}

func hashSessionToken(sessionToken string) string {
	sum := sha256.Sum256([]byte(sessionToken))
	return hex.EncodeToString(sum[:])
}
