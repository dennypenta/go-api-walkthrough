package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/dennypenta/go-api-walkthrough/domain"
	"github.com/dennypenta/go-api-walkthrough/pkg/log"
)

//go:generate mockery --name=UserService --dir=. --outpkg=mocks --filename=mock_user_service.go --output=./mocks --structname MockUserService
type UserService interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUserByID(ctx context.Context, id string) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, filter domain.UserFilter) (domain.PaginatedUserList, error)
}

type Handler struct {
	service UserService

	defaultLimit  int
	defaultOffset int
}

func NewHandler(service UserService) *Handler {
	return &Handler{
		service: service,

		defaultLimit:  10,
		defaultOffset: 0,
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

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeJson(w, ErrFailedMarshal, 400)
		return
	}

	user, err := h.service.CreateUser(r.Context(), user)
	if err != nil {
		handleError(r.Context(), err, w)
		return
	}

	writeJson(w, user, 200)
}

func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		handleError(r.Context(), err, w)
		return
	}

	writeJson(w, user, 200)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeJson(w, ErrFailedMarshal, 400)
		return
	}

	user, err := h.service.UpdateUser(r.Context(), user)
	if err != nil {
		handleError(r.Context(), err, w)
		return
	}

	writeJson(w, user, 200)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.DeleteUser(r.Context(), id); err != nil {
		handleError(r.Context(), err, w)
		return
	}

	w.WriteHeader(200)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit := h.defaultLimit
	offset := h.defaultOffset

	if limitStr != "" {
		limitV, err := strconv.Atoi(limitStr)
		// limit can't be 0
		if err == nil && limitV > 0 {
			limit = limitV
		}
	}
	if offsetStr != "" {
		offsetV, err := strconv.Atoi(offsetStr)
		// offset can't be negative
		if err == nil && offsetV >= 0 {
			offset = offsetV
		}
	}

	users, err := h.service.ListUsers(r.Context(), domain.UserFilter{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		handleError(r.Context(), err, w)
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

func handleError(ctx context.Context, err error, w http.ResponseWriter) {
	l := log.LoggerFromContext(ctx)

	switch {
	case errors.Is(err, domain.ErrUserExists):
		writeJson(w, ErrUserExists, 400)
	case errors.Is(err, domain.ErrUserNotFound):
		writeJson(w, ErrUserNotFound, 400)
	case errors.Is(err, domain.ErrInvalidUsername):
		writeJson(w, ErrInvalidUsername, 400)

	default:
		l.ErrorContext(ctx, "unhandled error", "err", err)
		writeJson(w, ErrUnknown, 500)
	}
}
