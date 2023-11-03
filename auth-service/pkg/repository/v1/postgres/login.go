package postgres

import (
	"context"
	"fmt"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/repository/v1"
)

const (
	selectUserInfo        = "SELECT username,password,cont_failed_login,role,locale FROM users WHERE username = $1"
	incFailedLoginCount   = "UPDATE users SET cont_failed_login = cont_failed_login + 1  WHERE username = $1"
	resetFailedLoginCount = "UPDATE users SET cont_failed_login = 0, last_login = NOW()   WHERE username = $1"
)

// UserInfo implements Database UserInfo function.
func (d *Default) UserInfo(ctx context.Context, userID string) (*v1.UserInfo, error) {
	ui := &v1.UserInfo{}
	if err := d.db.QueryRowContext(ctx, selectUserInfo, userID).
		Scan(&ui.UserID, &ui.Password, &ui.FailedLogins, &ui.Role, &ui.Locale); err != nil {
		return nil, err
	}
	return ui, nil
}

// IncreaseFailedLoginCount implements Database IncreaseFailedLoginCount function.
func (d *Default) IncreaseFailedLoginCount(ctx context.Context, userID string) error {
	result, err := d.db.ExecContext(ctx, incFailedLoginCount, userID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return fmt.Errorf("database - IncreaseFailedLoginCount- expected updated rows: 1, actual: %v", n)
	}

	return nil
}

// ResetLoginCount implements Database ResetLoginCount function.
func (d *Default) ResetLoginCount(ctx context.Context, userID string) error {
	result, err := d.db.ExecContext(ctx, resetFailedLoginCount, userID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return fmt.Errorf("database - ResetLoginCount - expected updated rows: 1, actual: %v", n)
	}

	return nil
}
