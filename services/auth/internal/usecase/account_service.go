package usecase

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/domain"
)

func (service *Service) ChangePassword(ctx context.Context, command ChangePasswordCommand) error {
	if len(command.NewPassword) < 8 {
		return ErrWeakPassword
	}

	account, err := service.authenticateByPassword(ctx, command.UserID, command.CurrentPassword)
	if err != nil {
		return err
	}

	passwordHash, err := service.passwordHasher.Hash(ctx, command.NewPassword)
	if err != nil {
		return err
	}

	return service.accountRepository.UpdatePassword(ctx, account.ID, passwordHash, service.clock.Now())
}

func (service *Service) ChangeEmail(ctx context.Context, command ChangeEmailCommand) (domain.Account, error) {
	command.NewEmail = normalizeEmail(command.NewEmail)
	if !isValidEmail(command.NewEmail) {
		return domain.Account{}, ErrInvalidEmail
	}

	account, err := service.authenticateByPassword(ctx, command.UserID, command.CurrentPassword)
	if err != nil {
		return domain.Account{}, err
	}

	return service.accountRepository.UpdateEmail(ctx, account.ID, command.NewEmail, service.clock.Now())
}

func (service *Service) PromoteToTrainer(ctx context.Context, command PromoteToTrainerCommand) (domain.Account, error) {
	account, err := service.accountRepository.GetByID(ctx, command.UserID)
	if err != nil {
		return domain.Account{}, err
	}

	if account.Role != domain.RoleClient {
		return domain.Account{}, domain.ErrAlreadyTrainer
	}

	return service.accountRepository.UpdateRole(ctx, account.ID, domain.RoleTrainer, service.clock.Now())
}

func (service *Service) LogoutAll(ctx context.Context, command LogoutAllCommand) error {
	if command.UserID <= 0 {
		return domain.ErrAccountNotFound
	}

	return service.sessionRepository.RevokeByUserID(ctx, command.UserID)
}

func (service *Service) DeleteAccount(ctx context.Context, command DeleteAccountCommand) error {
	account, err := service.authenticateByPassword(ctx, command.UserID, command.CurrentPassword)
	if err != nil {
		return err
	}

	return service.accountRepository.Delete(ctx, account.ID)
}

func (service *Service) authenticateByPassword(ctx context.Context, userID int64, password string) (domain.Account, error) {
	account, err := service.accountRepository.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return domain.Account{}, domain.ErrInvalidCredentials
		}

		return domain.Account{}, err
	}

	if err := service.passwordHasher.Compare(ctx, account.PasswordHash, password); err != nil {
		return domain.Account{}, domain.ErrInvalidCredentials
	}

	return account, nil
}
