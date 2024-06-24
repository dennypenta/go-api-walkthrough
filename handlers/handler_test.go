package handlers_test

import (
	"bytes"
	_ "embed"
	"net/http/httptest"
	"testing"

	"github.com/dennypenta/go-api-walkthrough/domain"
	"github.com/dennypenta/go-api-walkthrough/handlers"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/user.json
var userJson string

func TestCreateUserHandler(t *testing.T) {
	type testCase struct {
		name       string
		reqBody    []byte
		setupMocks func(m *MockUserService)

		expectedResp   string
		expectedStatus int
	}
	user := domain.User{ID: "8da80ba8-81c6-4336-bba3-ba8ea50541b0", Username: "test"}

	for _, tt := range []testCase{
		{
			name:    "valid request",
			reqBody: []byte(`{"username": "test"}`),
			setupMocks: func(m *MockUserService) {
				m.On("CreateUser", domain.User{Username: "test"}).Return(user, nil)
			},
			expectedResp:   userJson,
			expectedStatus: 200,
		},
		{
			name:    "invalid request",
			reqBody: []byte(`{"username": ""}`),
			setupMocks: func(m *MockUserService) {
				m.On("CreateUser", domain.User{}).Return(domain.User{}, domain.ErrInvalidUsername)
			},
			expectedResp:   `{"code":"invalid_username"}`,
			expectedStatus: 400,
		},
		{
			name:    "user exists",
			reqBody: []byte(`{"username": "test"}`),
			setupMocks: func(m *MockUserService) {
				m.On("CreateUser", domain.User{Username: "test"}).Return(domain.User{}, domain.ErrUserExists)
			},
			expectedResp:   `{"code":"user_exists"}`,
			expectedStatus: 400,
		},
		{
			name:    "failed marshal",
			reqBody: []byte(`{`),
			setupMocks: func(m *MockUserService) {
			},
			expectedResp:   `{"code":"failed_marshal"}`,
			expectedStatus: 400,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMockUserService(t)
			tt.setupMocks(m)

			h := handlers.NewHandler(m)
			req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(tt.reqBody))
			w := httptest.NewRecorder()
			h.CreateUser(req, w)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedResp, w.Body.String())
		})
	}
}
