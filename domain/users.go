package domain

import (
	"context"
)

//go:generate mockery --name=UserRepository --dir=. --outpkg=mocks --filename=mock_user_repository.go --output=./mocks --structname MockUserRepository
type UserRepository interface {
	CreateUser(ctx context.Context, user User) (User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
	UpdateUser(ctx context.Context, user User) (User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, filter UserFilter) ([]User, int, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user User) (User, error) {
	if err := user.Validate(); err != nil {
		return user, err
	}
	return s.repo.CreateUser(ctx, user)
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, user User) (User, error) {
	if err := user.Validate(); err != nil {
		return user, err
	}
	return s.repo.UpdateUser(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteUser(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context, filter UserFilter) (PaginatedUserList, error) {
	users, count, err := s.repo.ListUsers(ctx, filter)
	if err != nil {
		return PaginatedUserList{}, err
	}

	paginatedList := PaginatedUserList{
		Users:  users,
		Total:  count,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}
	paginatedList.EnrichHttpQueryLinks()
	return paginatedList, nil
}
