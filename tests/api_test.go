//go:build integration

package tests_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/stdlib"

	"github.com/dennypenta/go-api-walkthrough/domain"
	"github.com/dennypenta/go-api-walkthrough/handlers"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApi(t *testing.T) {
	// clean database on start
	db, err := sqlx.Open("pgx", "postgres://pguser:pgpass@localhost:5432/main?sslmode=disable")
	require.NoError(t, err)
	defer db.Close()
	db.Exec("TRUNCATE TABLE users")
	// clean database on end
	m, err := migrate.New("file://../migrations", "postgres://pguser:pgpass@localhost:5432/main?sslmode=disable")
	require.NoError(t, err)
	defer m.Down()

	c := http.DefaultClient

	baseUrl := "http://localhost:8080/v1"

	// create 30 users
	users := make([]domain.User, 0, 30)
	for i := 0; i < 30; i++ {
		payload := newCreateUserPayload(fmt.Sprintf("user-%d", i))
		resp, err := c.Post(baseUrl+"/users", "application/json", bytes.NewBuffer(payload))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var user domain.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		require.NoError(t, err)
		users = append(users, user)

		err = resp.Body.Close()
		require.NoError(t, err)

		// uuid v4 is 36 characters
		assert.Len(t, user.ID, 36)
	}

	// test invalid username on create
	resp, err := c.Post(baseUrl+"/users", "application/json", bytes.NewBuffer([]byte(`{"username": ""}`)))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	httpErr := handlers.Error{}
	err = json.NewDecoder(resp.Body).Decode(&httpErr)
	require.NoError(t, err)
	require.Equal(t, handlers.ErrInvalidUsername.Code, httpErr.Code)

	// test invalid username on update
	req, err := http.NewRequest("PUT", baseUrl+"/users/"+users[0].ID, bytes.NewBuffer([]byte(`{"username": ""}`)))
	require.NoError(t, err)
	resp, err = c.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	httpErr = handlers.Error{}
	err = json.NewDecoder(resp.Body).Decode(&httpErr)
	require.NoError(t, err)
	require.Equal(t, handlers.ErrInvalidUsername.Code, httpErr.Code)

	// update first user
	updatedUsername := "updated-user-0"
	req, err = http.NewRequest("PUT", baseUrl+"/users/"+users[0].ID, bytes.NewBuffer(newCreateUserPayload(updatedUsername)))
	require.NoError(t, err)
	resp, err = c.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&users[0])
	require.NoError(t, err)
	assert.Equal(t, updatedUsername, users[0].Username)

	// users list
	resp, err = http.Get(baseUrl + "/users?limit=10&offset=20")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	usersList := domain.PaginatedUserList{}
	err = json.NewDecoder(resp.Body).Decode(&usersList)
	require.NoError(t, err)
	require.Equal(t, 10, len(usersList.Users))
	assert.Equal(t, updatedUsername, usersList.Users[9].Username)

	// pagination
	assert.Equal(t, 30, usersList.Total)
	assert.Equal(t, 3, usersList.Pages)
	assert.Equal(t, 3, usersList.Page)
	assert.Equal(t, "", usersList.Next)
	assert.Equal(t, "limit=10&offset=10", usersList.Prev)

	// the rest users weren't updated
	userNum := 1
	for i := 8; i >= 0; i-- {
		username := fmt.Sprintf("user-%d", userNum)
		assert.Equal(t, username, usersList.Users[i].Username)
		userNum++
	}

	// delete the updated user
	req, err = http.NewRequest("DELETE", baseUrl+"/users/"+users[0].ID, nil)
	require.NoError(t, err)
	resp, err = c.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// test not found error
	resp, err = c.Get(baseUrl + "/users/" + users[0].ID)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	httpErr = handlers.Error{}
	err = json.NewDecoder(resp.Body).Decode(&httpErr)
	require.NoError(t, err)
	require.Equal(t, handlers.ErrUserNotFound.Code, httpErr.Code)

	// deleted user is not in the list
	resp, err = http.Get(baseUrl + "/users?limit=30&offset=0")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	usersList = domain.PaginatedUserList{}
	err = json.NewDecoder(resp.Body).Decode(&usersList)
	require.NoError(t, err)
	assert.Equal(t, 29, len(usersList.Users))
	assert.Equal(t, 29, usersList.Total)
	assert.Equal(t, 1, usersList.Pages)
	assert.Equal(t, 1, usersList.Page)
	assert.Equal(t, "", usersList.Next)
	assert.Equal(t, "", usersList.Prev)

	for _, u := range usersList.Users {
		assert.True(t, strings.HasPrefix(u.Username, "user-"))
	}
}

func newCreateUserPayload(username string) []byte {
	return []byte(fmt.Sprintf(`{"username": "%s"}`, username))
}
