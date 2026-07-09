package service

import (
	"context"

	"weather-accuracy/core-api/internal/domain"
)

func (s *Service) ListUsers(ctx context.Context, role domain.UserRole) ([]domain.User, error) {
	if err := s.RequireRole(role, domain.RoleAdmin); err != nil {
		return nil, err
	}
	users, err := s.deps.Users.List(ctx)
	for i := range users {
		users[i].PasswordHash = ""
	}
	return users, err
}

func (s *Service) UpdateUserRole(ctx context.Context, id string, role domain.UserRole, actor domain.UserRole) (domain.User, error) {
	if err := s.RequireRole(actor, domain.RoleAdmin); err != nil {
		return domain.User{}, err
	}
	user, err := s.deps.Users.UpdateRole(ctx, id, role)
	user.PasswordHash = ""
	return user, err
}
