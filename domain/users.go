package domain

type UserRepository interface {
	CreateUser(user User) (User, error)
	GetUserByID(id string) (User, error)
	UpdateUser(user User) error
	DeleteUser(id string) error
	ListUsers() ([]User, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(user User) (User, error) {
	return s.repo.CreateUser(user)
}

func (s *UserService) GetUserByID(id string) (User, error) {
	return s.repo.GetUserByID(id)
}

func (s *UserService) UpdateUser(user User) error {
	return s.repo.UpdateUser(user)
}

func (s *UserService) DeleteUser(id string) error {
	return s.repo.DeleteUser(id)
}

func (s *UserService) ListUsers() ([]User, error) {
	return s.repo.ListUsers()
}
