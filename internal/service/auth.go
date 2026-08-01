package service

import (
	"errors"
	"net/http"

	"forum/internal/apperror"
	"forum/internal/auth"
	"forum/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db                *gorm.DB
	tokens            *auth.TokenManager
	dummyPasswordHash []byte
}

func NewAuthService(db *gorm.DB, tokens *auth.TokenManager) *AuthService {
	dummyPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("forum-dummy-password"), bcrypt.DefaultCost)
	return &AuthService{db: db, tokens: tokens, dummyPasswordHash: dummyPasswordHash}
}

type RegisterInput struct {
	Username string
	Name     string
	Password string
	Role     string
}

type UserView struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type LoginResult struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int64    `json:"expires_in"`
	User        UserView `json:"user"`
}
//注册*加密后存入mysql
func (s *AuthService) Register(input RegisterInput) (UserView, error) {
	var count int64
	if err := s.db.Model(&model.User{}).Where("username = ?", input.Username).Count(&count).Error; err != nil {
		return UserView{}, err
	}
	if count > 0 {
		return UserView{}, apperror.New(http.StatusConflict, http.StatusConflict, "用户名已存在")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return UserView{}, err
	}
	user := model.User{
		Username:     input.Username,
		Name:         input.Name,
		PasswordHash: string(passwordHash),
		Role:         input.Role,
	}
	if err := s.db.Create(&user).Error; err != nil {
		var existing int64
		if countErr := s.db.Model(&model.User{}).Where("username = ?", input.Username).Count(&existing).Error; countErr == nil && existing > 0 {
			return UserView{}, apperror.New(http.StatusConflict, http.StatusConflict, "用户名已存在")
		}
		return UserView{}, err
	}
	return toUserView(user), nil
}

func (s *AuthService) Login(username, password string) (LoginResult, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = bcrypt.CompareHashAndPassword(s.dummyPasswordHash, []byte(password))
			return LoginResult{}, apperror.New(http.StatusUnauthorized, http.StatusUnauthorized, "账号或密码错误")
		}
		return LoginResult{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return LoginResult{}, apperror.New(http.StatusUnauthorized, http.StatusUnauthorized, "账号或密码错误")
	}
	token, expiresIn, err := s.tokens.Create(user)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        toUserView(user),
	}, nil
}

func toUserView(user model.User) UserView {
	return UserView{ID: user.ID, Username: user.Username, Name: user.Name, Role: user.Role}
}
