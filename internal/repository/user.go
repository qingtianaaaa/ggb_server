package repository

import (
	"ggb_server/internal/app/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	Generic[model.User]
	GetByUsername(db *gorm.DB, username string) (*model.User, error)
	GetByEmail(db *gorm.DB, email string) (*model.User, error)
	GetByInviteCode(db *gorm.DB, inviteCode string) (*model.User, error)
	UpdateLoginInfo(db *gorm.DB, userID uint) error
}

type UserRepo struct {
	GenericImpl[model.User]
}

func NewUserRepository() UserRepository {
	return &UserRepo{}
}

func (r *UserRepo) GetByUsername(db *gorm.DB, username string) (*model.User, error) {
	var user model.User
	err := db.Where("username = ? AND is_del = ?", username, false).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetByEmail(db *gorm.DB, email string) (*model.User, error) {
	var user model.User
	err := db.Where("email = ? AND is_del = ?", email, false).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetByInviteCode(db *gorm.DB, inviteCode string) (*model.User, error) {
	var user model.User
	err := db.Where("invite_code = ? AND is_del = ?", inviteCode, false).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) UpdateLoginInfo(db *gorm.DB, userID uint) error {
	return db.Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": gorm.Expr("NOW()"),
			"login_count":   gorm.Expr("login_count + 1"),
		}).Error
}
