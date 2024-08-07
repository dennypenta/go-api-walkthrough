package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/dennypenta/go-api-walkthrough/domain"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
	sq sq.StatementBuilderType
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	query, args, err := r.sq.Insert("users").
		Columns("username").
		Values(user.Username).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return user, fmt.Errorf("CreateUser: failed to build query: %w", err)
	}

	var id string
	if err := r.db.QueryRowxContext(ctx, query, args...).Scan(&id); err != nil {
		return user, fmt.Errorf("CreateUser: failed to insert user: %w", err)
	}

	user.ID = id
	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (domain.User, error) {
	var user domain.User
	query, args, err := r.sq.Select("username").
		From("users").
		Where(sq.Eq{"id": id, "deletedAt": nil}).
		ToSql()
	if err != nil {
		return user, fmt.Errorf("GetUserByID: failed to build query: %w", err)
	}

	err = r.db.QueryRowxContext(ctx, query, args...).Scan(&user.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, domain.ErrUserNotFound
		}
		return user, err
	}

	user.ID = id
	return user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	query, args, err := r.sq.Update("users").
		Set("username", user.Username).Set("updatedAt", sq.Expr("now()")).
		Where(sq.Eq{"id": user.ID}).
		ToSql()
	if err != nil {
		return user, fmt.Errorf("UpdateUser: failed to build query: %w", err)
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return user, fmt.Errorf("UpdateUser: failed to update user: %w", err)
	}

	affectedAmount, err := res.RowsAffected()
	if err != nil {
		return user, fmt.Errorf("UpdateUser: failed to get RowsAffected: %w", err)
	}
	if affectedAmount == 0 {
		return user, domain.ErrUserNotFound
	}

	return user, nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	query, args, err := r.sq.Update("users").
		Set("deletedAt", sq.Expr("now()")).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("DeleteUser: failed to build query: %w", err)
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("DeleteUser: failed to delete user: %w", err)
	}

	affectedAmount, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteUser: failed to get RowsAffected: %w", err)
	}
	if affectedAmount == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) ListUsers(ctx context.Context, filter domain.UserFilter) ([]domain.User, int, error) {
	var count int
	var users []domain.User
	subQ := r.sq.Select("id", "username", "COUNT(*) OVER () AS total").
		Where(sq.Eq{"deletedAt": nil}).Limit(uint64(filter.Limit)).
		Offset(uint64(filter.Offset)).
		OrderBy("createdAt DESC").From("users")
	query, args, err := r.sq.Select("id", "username", "total").
		FromSelect(subQ, "sub").
		ToSql()
	if err != nil {
		return users, 0, fmt.Errorf("ListUsers: failed to build query: %w", err)
	}

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return users, 0, fmt.Errorf("ListUsers: failed to list users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Username, &count); err != nil {
			return users, 0, fmt.Errorf("ListUsers: failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, count, nil
}
