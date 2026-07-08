package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"weather-accuracy/core-api/internal/domain"
)

type Claims struct {
	UserID string          `json:"uid"`
	Role   domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func (s *Service) Login(ctx context.Context, login, password string) (string, domain.User, error) {
	user, err := s.deps.Users.FindByLogin(ctx, login)
	if err != nil {
		return "", domain.User{}, domain.ErrInvalidCredentials
	}
	if !user.IsActive {
		return "", domain.User{}, domain.ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", domain.User{}, domain.ErrInvalidCredentials
	}
	_ = s.deps.Users.TouchLastSeen(ctx, user.ID)
	token, err := s.issueToken(user)
	if err != nil {
		return "", domain.User{}, err
	}
	user.PasswordHash = ""
	return token, user, nil
}

func (s *Service) CurrentUser(ctx context.Context, id string) (domain.User, error) {
	user, err := s.deps.Users.FindByID(ctx, id)
	user.PasswordHash = ""
	return user, err
}

func (s *Service) ParseToken(tokenString string) (Claims, error) {
	claims := Claims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(s.deps.JWTSecret), nil
	}, jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil || !token.Valid {
		return Claims{}, domain.ErrInvalidCredentials
	}
	return claims, nil
}

func (s *Service) issueToken(user domain.User) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.deps.TokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.deps.JWTSecret))
}
