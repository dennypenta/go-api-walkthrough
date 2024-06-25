package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dennypenta/go-api-walkthrough/domain"
)

//go:generate mockery --name=UserService --dir=. --outpkg=mocks --filename=mock_user_service.go --output=./mocks --structname MockUserService
type UserService interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUserByID(ctx context.Context, id string) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context) ([]domain.User, error)
}

type Handler struct {
	service UserService
}

func NewHandler(service UserService) *Handler {
	return &Handler{
		service: service,
	}
}

type Error struct {
	Code string                 `json:"code"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

var (
	ErrUnknown = Error{
		Code: "unknown",
	}

	ErrFailedMarshal = Error{
		Code: "failed_marshal",
	}
	ErrInvalidUsername = Error{
		Code: "invalid_username",
	}
	ErrUserNotFound = Error{
		Code: "user_not_found",
	}
	ErrUserExists = Error{
		Code: "user_exists",
	}
)

func (h *Handler) CreateUser(r *http.Request, w http.ResponseWriter) {
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeJson(w, ErrFailedMarshal, 400)
		return
	}

	user, err := h.service.CreateUser(r.Context(), user)
	if err != nil {
		handleError(err, w)
		return
	}

	writeJson(w, user, 200)
}

func (h *Handler) GetUserByID(r *http.Request, w http.ResponseWriter) {
	id := r.PathValue("id")
	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		handleError(err, w)
		return
	}

	writeJson(w, user, 200)
}

func (h *Handler) UpdateUser(r *http.Request, w http.ResponseWriter) {
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeJson(w, ErrFailedMarshal, 400)
		return
	}

	user, err := h.service.UpdateUser(r.Context(), user)
	if err != nil {
		handleError(err, w)
		return
	}

	writeJson(w, user, 200)
}

func (h *Handler) DeleteUser(r *http.Request, w http.ResponseWriter) {
	id := r.PathValue("id")
	if err := h.service.DeleteUser(r.Context(), id); err != nil {
		handleError(err, w)
		return
	}

	w.WriteHeader(200)
}

func (h *Handler) ListUsers(r *http.Request, w http.ResponseWriter) {
	users, err := h.service.ListUsers(r.Context())
	if err != nil {
		handleError(err, w)
		return
	}

	writeJson(w, users, 200)
}

func writeJson(w http.ResponseWriter, v interface{}, status int) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func handleError(err error, w http.ResponseWriter) {
	switch {
	case errors.Is(err, domain.ErrUserExists):
		writeJson(w, ErrUserExists, 400)
	case errors.Is(err, domain.ErrUserNotFound):
		writeJson(w, ErrUserNotFound, 400)
	case errors.Is(err, domain.ErrInvalidUsername):
		writeJson(w, ErrInvalidUsername, 400)

	default:
		writeJson(w, ErrUnknown, 500)
	}
}
