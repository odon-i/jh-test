package auth

import (
	"errors"
	"fmt"
	"time"

	"forum/internal/config"
	"forum/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
//管理token
type TokenManager struct {
	secret    []byte
	issuer    string
	expiresIn time.Duration
	now       func() time.Time
}

func NewTokenManager(cfg config.JWTConfig) *TokenManager {
	return &TokenManager{secret: []byte(cfg.Secret), issuer: cfg.Issuer, expiresIn: cfg.ExpiresIn, now: time.Now}
}

func (m *TokenManager) Create(user model.User) (string, int64, error) {
	now := m.now()
	expiresAt := now.Add(m.expiresIn)
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	return token, int64(m.expiresIn.Seconds()), err
}

func (m *TokenManager) Parse(raw string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %s", token.Method.Alg())
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
