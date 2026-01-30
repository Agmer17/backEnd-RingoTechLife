package user

import (
	"backEnd-RingoTechLife/internal/common/model"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepositoryInterface interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (model.User, error)
	GetByEmailOrPhone(ctx context.Context, email string, phoneNumber string) (model.User, error)
	Update(ctx context.Context, user *model.User) (*model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	IsUserExists(ctx context.Context, email string, phone string) (bool, error)
}

type userRepositoryImpl struct {
	db *pgxpool.Pool
}

func (r *userRepositoryImpl) Create(
	ctx context.Context,
	tx pgx.Tx,
	user *model.User,
) (*model.User, error) {

	query := `
		INSERT INTO users 
			(full_name, email, phone_number, password)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query,
			user.FullName,
			user.Email,
			user.PhoneNumber,
			user.Password,
		).Scan(&user.ID, &user.CreatedAt)
	})

	if err != nil {
		return nil, fmt.Errorf("create user failed: %w", err)
	}

	return user, nil
}

func (r *userRepositoryImpl) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (model.User, error) {

	query := `
		SELECT id, full_name, email, phone_number, password, role, profile_picture, created_at
		FROM users
		WHERE id = $1
		limit 1
	`

	var u model.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.FullName,
		&u.Email,
		&u.PhoneNumber,
		&u.Password,
		&u.Role,
		&u.ProfilePicture,
		&u.CreatedAt,
	)

	if err != nil {
		return model.User{}, err
	}

	return u, nil
}

func (r *userRepositoryImpl) Update(
	ctx context.Context,
	user *model.User,
) (*model.User, error) {

	query := `
		UPDATE users 
		SET full_name = $1,
		    email = $2,
		    phone_number = $3,
		    role = $4,
		    profile_picture = $5
		WHERE id = $6
		RETURNING created_at
	`

	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query,
			user.FullName,
			user.Email,
			user.PhoneNumber,
			user.Role,
			user.ProfilePicture,
			user.ID,
		).Scan(&user.CreatedAt)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (r *userRepositoryImpl) Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	res, err := tx.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("user with id %s not found", id)
	}

	return nil
}

func (r *userRepositoryImpl) GetByEmailOrPhone(
	ctx context.Context,
	email string,
	phoneNumber string,
) (model.User, error) {

	query := `
		SELECT id, full_name, email, phone_number, password, role, profile_picture, created_at
		FROM users
		WHERE email = $1 OR phone_number = $2
		LIMIT 1
	`

	var u model.User
	err := r.db.QueryRow(ctx, query, email, phoneNumber).Scan(
		&u.ID,
		&u.FullName,
		&u.Email,
		&u.PhoneNumber,
		&u.Password,
		&u.Role,
		&u.ProfilePicture,
		&u.CreatedAt,
	)

	if err != nil {
		return model.User{}, err
	}

	return u, nil
}

func (r *userRepositoryImpl) IsUserExists(
	ctx context.Context,
	email string,
	phone string,
) (bool, error) {

	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE email = $1 OR phone_number = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, email, phone).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
