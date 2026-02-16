package review

import (
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/common/model"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ---- Sentinel Errors ----

var (
	ErrReviewNotFound        = errors.New("review not found")
	ErrReviewAlreadyExists   = errors.New("user already reviewed this product")
	ErrReviewUserNotFound    = errors.New("user not found")
	ErrReviewProductNotFound = errors.New("product not found")
)

const reviewDetailQuery = `
	SELECT
		r.id,
		r.product_id,
		r.rating,
		r.comment,
		r.created_at,
		u.id          AS user_id,
		u.full_name   AS user_full_name,
		u.profile_picture
	FROM reviews r
	JOIN users u ON u.id = r.user_id
`

// ---- View Models ----

// ---- Interface ----

type ReviewRepository interface {
	// Plain CRUD
	GetByID(ctx context.Context, id uuid.UUID) (*model.Review, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]*model.Review, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Review, error)
	Create(ctx context.Context, review *model.Review) (*model.Review, error)
	Update(ctx context.Context, review *model.Review) (*model.Review, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Detail (JOIN dengan tabel users)
	GetAllDetails(ctx context.Context) ([]*dto.ReviewDetail, error)
	GetDetailByID(ctx context.Context, id uuid.UUID) (*dto.ReviewDetail, error)
	GetDetailsByUserID(ctx context.Context, userID uuid.UUID) ([]*dto.ReviewDetail, error)
	GetAllDetailsProdId(ctx context.Context, productId uuid.UUID) ([]dto.ReviewDetail, error)

	// buat yang join sama table products, nanti di bikin aja methodnya
	// di products repo jangan disini
	// soalnya review tuh di own sama products.
}

// ---- Struct & Constructor ----

type ReviewRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewReviewRepository(pool *pgxpool.Pool) *ReviewRepositoryImpl {
	return &ReviewRepositoryImpl{
		db: pool,
	}
}

// ---- Read Methods ----

func (r *ReviewRepositoryImpl) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.Review, error) {

	query := `
		SELECT id, product_id, user_id, rating, comment, created_at
		FROM reviews
		WHERE id = $1
		LIMIT 1
	`

	var review model.Review
	err := r.db.QueryRow(ctx, query, id).Scan(
		&review.ID,
		&review.ProductID,
		&review.UserID,
		&review.Rating,
		&review.Comment,
		&review.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReviewNotFound
		}
		return nil, fmt.Errorf("get review by id failed: %w", err)
	}

	return &review, nil
}

func (r *ReviewRepositoryImpl) GetByProductID(
	ctx context.Context,
	productID uuid.UUID,
) ([]*model.Review, error) {

	query := `
		SELECT id, product_id, user_id, rating, comment, created_at
		FROM reviews
		WHERE product_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("get reviews by product_id failed: %w", err)
	}
	defer rows.Close()

	var reviews []*model.Review
	for rows.Next() {
		var review model.Review
		err := rows.Scan(
			&review.ID,
			&review.ProductID,
			&review.UserID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan review row failed: %w", err)
		}
		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return reviews, nil
}

func (r *ReviewRepositoryImpl) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]*model.Review, error) {

	query := `
		SELECT id, product_id, user_id, rating, comment, created_at
		FROM reviews
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get reviews by user_id failed: %w", err)
	}
	defer rows.Close()

	var reviews []*model.Review
	for rows.Next() {
		var review model.Review
		err := rows.Scan(
			&review.ID,
			&review.ProductID,
			&review.UserID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan review row failed: %w", err)
		}
		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return reviews, nil
}

// ---- Detail Methods (JOIN dengan tabel users) ----

// reviewScanner adalah interface minimal yang dipenuhi oleh pgx.Row maupun pgx.Rows,
// sehingga scanReviewDetail bisa dipakai untuk keduanya tanpa duplikasi kode.
type reviewScanner interface {
	Scan(dest ...any) error
}

// scanReviewDetail adalah helper untuk scan satu row ReviewDetail.
func scanReviewDetail(row reviewScanner) (*dto.ReviewDetail, error) {
	var d dto.ReviewDetail
	err := row.Scan(
		&d.ID,
		&d.ProductID,
		&d.Rating,
		&d.Comment,
		&d.CreatedAt,
		&d.User.ID,
		&d.User.FullName,
		&d.User.ProfilePicture,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *ReviewRepositoryImpl) GetAllDetails(
	ctx context.Context,
) ([]*dto.ReviewDetail, error) {

	query := reviewDetailQuery + `ORDER BY r.created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all review details failed: %w", err)
	}
	defer rows.Close()

	var results []*dto.ReviewDetail
	for rows.Next() {
		d, err := scanReviewDetail(rows)
		if err != nil {
			return nil, fmt.Errorf("scan review detail row failed: %w", err)
		}
		results = append(results, d)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, nil
}

func (r *ReviewRepositoryImpl) GetDetailByID(
	ctx context.Context,
	id uuid.UUID,
) (*dto.ReviewDetail, error) {

	query := reviewDetailQuery + `WHERE r.id = $1 LIMIT 1`

	d, err := scanReviewDetail(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReviewNotFound
		}
		return nil, fmt.Errorf("get review detail by id failed: %w", err)
	}

	return d, nil
}

func (r *ReviewRepositoryImpl) GetDetailsByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]*dto.ReviewDetail, error) {

	query := reviewDetailQuery + `WHERE r.user_id = $1 ORDER BY r.created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get review details by user_id failed: %w", err)
	}
	defer rows.Close()

	var results []*dto.ReviewDetail
	for rows.Next() {
		d, err := scanReviewDetail(rows)
		if err != nil {
			return nil, fmt.Errorf("scan review detail row failed: %w", err)
		}
		results = append(results, d)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, nil
}

func (r *ReviewRepositoryImpl) GetAllDetailsProdId(ctx context.Context, productId uuid.UUID) ([]dto.ReviewDetail, error) {
	query := reviewDetailQuery + `WHERE r.product_id = $1 ORDER BY r.created_at DESC`
	rows, err := r.db.Query(ctx, query, productId)
	if err != nil {
		return []dto.ReviewDetail{}, fmt.Errorf("get review details by product id failed: %w", err)
	}

	defer rows.Close()

	results := make([]dto.ReviewDetail, 0)
	for rows.Next() {
		d, err := scanReviewDetail(rows)
		if err != nil {
			return []dto.ReviewDetail{}, fmt.Errorf("scan review detail row failed: %w", err)
		}
		results = append(results, *d)
	}

	if err = rows.Err(); err != nil {
		return []dto.ReviewDetail{}, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, nil

}

// ---- Write Methods ----

func (r *ReviewRepositoryImpl) Create(
	ctx context.Context,
	review *model.Review,
) (*model.Review, error) {

	query := `
		INSERT INTO reviews
			(product_id, user_id, rating, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query,
			review.ProductID,
			review.UserID,
			review.Rating,
			review.Comment,
		).Scan(&review.ID, &review.CreatedAt)
	})

	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				// unique_user_product_review constraint
				if pgErr.ConstraintName == "unique_user_product_review" {
					return nil, ErrReviewAlreadyExists
				}
			}
			if pgErr.Code == "23503" {
				switch pgErr.ConstraintName {
				case "reviews_user_id_fkey":
					return nil, ErrReviewUserNotFound
				case "reviews_product_id_fkey":
					return nil, ErrReviewProductNotFound
				}
			}
		}
		return nil, fmt.Errorf("create review failed: %w", err)
	}

	return review, nil
}

func (r *ReviewRepositoryImpl) Update(
	ctx context.Context,
	review *model.Review,
) (*model.Review, error) {

	query := `
		UPDATE reviews
		SET rating  = $1,
		    comment = $2
		WHERE id = $3
		RETURNING product_id, user_id, created_at
	`

	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query,
			review.Rating,
			review.Comment,
			review.ID,
		).Scan(
			&review.ProductID,
			&review.UserID,
			&review.CreatedAt,
		)
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReviewNotFound
		}
		return nil, fmt.Errorf("update review failed: %w", err)
	}

	return review, nil
}

func (r *ReviewRepositoryImpl) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	query := `DELETE FROM reviews WHERE id = $1`

	var err error
	pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		res, execErr := tx.Exec(ctx, query, id)
		if execErr != nil {
			err = fmt.Errorf("delete review failed: %w", execErr)
			return execErr
		}

		if res.RowsAffected() == 0 {
			err = ErrReviewNotFound
			return ErrReviewNotFound
		}

		return nil
	})

	return err
}
