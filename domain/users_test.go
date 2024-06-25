package domain_test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/dennypenta/go-api-walkthrough/domain"
	"github.com/dennypenta/go-api-walkthrough/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateUserHandler(t *testing.T) {
	type testCase struct {
		name string

		input      domain.User
		setupMocks func(m *mocks.MockUserRepository)

		expectedResp domain.User
		expectedErr  error
	}

	user := domain.User{Username: "test"}
	outputUser := domain.User{ID: "8da80ba8-81c6-4336-bba3-ba8ea50541b0", Username: "test"}

	for _, tt := range []testCase{
		{
			name:  "valid creation",
			input: user,
			setupMocks: func(m *mocks.MockUserRepository) {
				m.On("CreateUser", mock.Anything, user).Return(outputUser, nil)
			},
			expectedResp: outputUser,
		},
		{
			name: "invalid username",
			setupMocks: func(m *mocks.MockUserRepository) {
			},
			expectedErr: domain.ErrInvalidUsername,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks.NewMockUserRepository(t)
			tt.setupMocks(m)
			service := domain.NewUserService(m)

			ctx := context.Background()
			res, err := service.CreateUser(ctx, tt.input)

			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Equal(t, res, tt.expectedResp)
		})
	}
}
