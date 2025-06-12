package repo

import (
	"context"

	gmodel "github.com/pilinux/gorest/database/model"
	"gorm.io/gorm"
)

// AuthRepo provides methods for authentication-related database operations.
type AuthRepo struct {
	db *gorm.DB
}

// NewAuthRepo returns a new AuthRepo with the given database connection.
func NewAuthRepo(conn *gorm.DB) *AuthRepo {
	return &AuthRepo{
		db: conn,
	}
}

// TempEmailRepo provides methods for updating user's email address.
type TempEmailRepo struct {
	db *gorm.DB
}

// NewTempEmailRepo returns a new TempEmailRepo with the given database connection.
func NewTempEmailRepo(conn *gorm.DB) *TempEmailRepo {
	return &TempEmailRepo{
		db: conn,
	}
}

// TwoFARepo provides methods for two-factor authentication.
type TwoFARepo struct {
	db *gorm.DB
}

// NewTwoFARepo returns a new TwoFA with the given database connection.
func NewTwoFARepo(conn *gorm.DB) *TwoFARepo {
	return &TwoFARepo{
		db: conn,
	}
}

// TwoFABackupRepo provides methods for two-factor authentication backup codes.
type TwoFABackupRepo struct {
	db *gorm.DB
}

// NewTwoFABackupRepo returns a new TwoFABackup with the given database connection.
func NewTwoFABackupRepo(conn *gorm.DB) *TwoFABackupRepo {
	return &TwoFABackupRepo{
		db: conn,
	}
}

// DeleteAuth deletes an authentication record by authID.
func (r *AuthRepo) DeleteAuth(ctx context.Context, authID uint64) error {
	return r.db.WithContext(ctx).Where("auth_id = ?", authID).Delete(&gmodel.Auth{}).Error
}

// DeleteTempEmail deletes a temporary email record by authID.
func (r *TempEmailRepo) DeleteTempEmail(ctx context.Context, authID uint64) error {
	return r.db.WithContext(ctx).Where("id_auth = ?", authID).Delete(&gmodel.TempEmail{}).Error
}

// DeleteTwoFA deletes a two-factor authentication record by authID.
func (r *TwoFARepo) DeleteTwoFA(ctx context.Context, authID uint64) error {
	return r.db.WithContext(ctx).Where("id_auth = ?", authID).Delete(&gmodel.TwoFA{}).Error
}

// DeleteTwoFABackup deletes a two-factor authentication backup record by authID.
func (r *TwoFABackupRepo) DeleteTwoFABackup(ctx context.Context, authID uint64) error {
	return r.db.WithContext(ctx).Where("id_auth = ?", authID).Delete(&gmodel.TwoFABackup{}).Error
}
