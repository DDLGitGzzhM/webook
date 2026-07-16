package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserNotFound       = errors.New("邮箱不存在")
)

type UserDao interface {
	Insert(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)
	FindById(ctx context.Context, id int64) (*User, error)
}

type GormUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *GormUserDAO {
	return &GormUserDAO{
		db: db,
	}
}
func (dao *GormUserDAO) Insert(ctx context.Context, u *User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(u).Error
	var mySqlErr *mysql.MySQLError
	if errors.As(err, &mySqlErr) {
		const uniqueConflicts = 1062
		if mySqlErr.Number == uniqueConflicts {
			return ErrUserDuplicateEmail
		}
	}
	return nil
}

func (dao *GormUserDAO) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &u, err
}

func (dao *GormUserDAO) FindByPhone(ctx context.Context, phone string) (*User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &u, err
}

func (dao *GormUserDAO) FindById(ctx context.Context, id int64) (*User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &u, err
}

type User struct {
	Id       int64          `gorm:"primaryKey,autoIncrement"`
	Email    sql.NullString `gorm:"unique"`
	Password string
	Phone    sql.NullString `gorm:"unique"` // 唯一索引允许有多个空值

	Ctime int64
	Utime int64
}
