package category

import (
	"backEnd-RingoTechLife/internal/common/model"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepositoryInterface interface {
	Create(ctx context.Context, category *model.Category) (*model.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (model.Category, error)
	GetBySlug(ctx context.Context, slug string) (model.Category, error)
	GetByName(ctx context.Context, name string) (model.Category, error)
	Update(ctx context.Context, category *model.Category) (*model.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetAllCategories(ctx context.Context) ([]model.Category, error)
	ExistsBySlug(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error)
	ExistsByName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error)
	ExistByNameOrSlug(ctx context.Context, name string, slug string, excludeID *uuid.UUID) (bool, error)
	ExistsById(ctx context.Context, id uuid.UUID) (bool, error)
}

type CategoryRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepositoryImpl {
	return &CategoryRepositoryImpl{
		db: pool,
	}
}

func (r *CategoryRepositoryImpl) Create(
	ctx context.Context,
	category *model.Category,
) (*model.Category, error) {

	query := `
        INSERT INTO categories 
            (name, slug, description)
        VALUES ($1, $2, $3)
        RETURNING id, created_at
    `

	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query,
			category.Name,
			category.Slug,
			category.Description,
		).Scan(&category.ID, &category.CreatedAt)
	})

	if err != nil {
		return nil, err
	}

	return category, nil
}

func (r *CategoryRepositoryImpl) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (model.Category, error) {

	query := `
        SELECT id, name, slug, description, created_at
        FROM categories
        WHERE id = $1
        LIMIT 1
    `

	var c model.Category
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID,
		&c.Name,
		&c.Slug,
		&c.Description,
		&c.CreatedAt,
	)

	if err != nil {
		return model.Category{}, err
	}

	return c, nil
}

func (r *CategoryRepositoryImpl) GetBySlug(
	ctx context.Context,
	slug string,
) (model.Category, error) {

	query := `
        SELECT id, name, slug, description, created_at
        FROM categories
        WHERE slug = $1
        LIMIT 1
    `

	var c model.Category
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&c.ID,
		&c.Name,
		&c.Slug,
		&c.Description,
		&c.CreatedAt,
	)

	if err != nil {
		return model.Category{}, err
	}

	return c, nil
}

func (r *CategoryRepositoryImpl) GetByName(
	ctx context.Context,
	name string,
) (model.Category, error) {

	query := `
        SELECT id, name, slug, description, created_at
        FROM categories
        WHERE name = $1
        LIMIT 1
    `

	var c model.Category
	err := r.db.QueryRow(ctx, query, name).Scan(
		&c.ID,
		&c.Name,
		&c.Slug,
		&c.Description,
		&c.CreatedAt,
	)

	if err != nil {
		return model.Category{}, err
	}

	return c, nil
}

func (r *CategoryRepositoryImpl) Update(
	ctx context.Context,
	category *model.Category,
) (*model.Category, error) {

	query := `
        UPDATE categories 
        SET name = $1,
            slug = $2,
            description = $3
        WHERE id = $4
        RETURNING created_at
    `

	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query,
			category.Name,
			category.Slug,
			category.Description,
			category.ID,
		).Scan(&category.CreatedAt)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

func (r *CategoryRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM categories WHERE id = $1`

	var err error
	pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		res, execErr := tx.Exec(ctx, query, id)
		if execErr != nil {
			err = execErr
			return execErr
		}

		if res.RowsAffected() == 0 {
			err = fmt.Errorf("category with id %s not found", id)
			return err
		}

		return nil
	})

	return err
}

func (r *CategoryRepositoryImpl) GetAllCategories(ctx context.Context) ([]model.Category, error) {
	query := `
        SELECT id, name, slug, description, created_at
        FROM categories
        ORDER BY created_at DESC
    `

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Slug,
			&c.Description,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *CategoryRepositoryImpl) ExistsBySlug(
	ctx context.Context,
	slug string,
	excludeID *uuid.UUID,
) (bool, error) {

	var (
		query string
		args  []any
	)

	if excludeID != nil {
		query = `
            SELECT EXISTS (
                SELECT 1
                FROM categories
                WHERE slug = $1
                  AND id != $2
            )
        `
		args = []any{slug, *excludeID}
	} else {
		query = `
            SELECT EXISTS (
                SELECT 1
                FROM categories
                WHERE slug = $1
            )
        `
		args = []any{slug}
	}

	var exists bool
	err := r.db.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *CategoryRepositoryImpl) ExistsByName(
	ctx context.Context,
	name string,
	excludeID *uuid.UUID,
) (bool, error) {

	var (
		query string
		args  []any
	)

	if excludeID != nil {
		query = `
            SELECT EXISTS (
                SELECT 1
                FROM categories
                WHERE name = $1
                  AND id != $2
            )
        `
		args = []any{name, *excludeID}
	} else {
		query = `
            SELECT EXISTS (
                SELECT 1
                FROM categories
                WHERE name = $1
            )
        `
		args = []any{name}
	}

	var exists bool
	err := r.db.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *CategoryRepositoryImpl) ExistByNameOrSlug(ctx context.Context, name string, slug string, excludeID *uuid.UUID) (bool, error) {

	var (
		query string
		args  []any
	)

	if excludeID != nil {
		query = `
            SELECT EXISTS (
                SELECT 1
                FROM categories
                WHERE name = $1 OR slug = $2
                  AND id != $3
            )
        `
		args = []any{name, slug, *excludeID}
	} else {
		query = `
            SELECT EXISTS (
                SELECT 1
                FROM categories
                WHERE name = $1 OR slug = $2
            )
        `
		args = []any{name, slug}
	}

	var exists bool
	err := r.db.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil

}

func (r *CategoryRepositoryImpl) ExistsById(
	ctx context.Context,
	id uuid.UUID,
) (bool, error) {

	var (
		query string
		args  []any
	)

	query = `
            SELECT EXISTS (
                SELECT 1
                FROM categories
                WHERE id = $1
            )
        `
	args = []any{id}

	var exists bool
	err := r.db.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
