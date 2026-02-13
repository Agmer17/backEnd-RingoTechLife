package productimage

import (
	"backEnd-RingoTechLife/internal/common/model"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrrNotFound = errors.New("NO image found!")

// ProductImageRepoInterface defines the interface for product image repository
type ProductImageRepoInterface interface {
	// Single operations
	Create(ctx context.Context, image *model.ProductImage) (*model.ProductImage, error)
	GetByID(ctx context.Context, id uuid.UUID) (model.ProductImage, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]model.ProductImage, error)
	Update(ctx context.Context, image *model.ProductImage) (*model.ProductImage, error)
	Delete(ctx context.Context, productID uuid.UUID, imageID uuid.UUID) error

	// Bulk operations
	CreateBulk(ctx context.Context, images []*model.ProductImage) ([]*model.ProductImage, error)
	DeleteBulk(ctx context.Context, productID uuid.UUID, imageIDs []uuid.UUID) error
	UpdateBulk(ctx context.Context, images []*model.ProductImage) ([]*model.ProductImage, error)

	GetAllByIDs(ctx context.Context, ids []uuid.UUID) ([]model.ProductImage, error)
}

// ProductImageRepoImpl implements ProductImageRepoInterface
type ProductImageRepoImpl struct {
	db *pgxpool.Pool
}

// NewProductImageRepository creates a new instance of ProductImageRepoImpl
func NewProductImageRepository(pool *pgxpool.Pool) *ProductImageRepoImpl {
	return &ProductImageRepoImpl{
		db: pool,
	}
}

// Create inserts a new product image
func (r *ProductImageRepoImpl) Create(
	ctx context.Context,
	image *model.ProductImage,
) (*model.ProductImage, error) {
	query := `
		INSERT INTO product_images 
			(product_id, image_url, is_primary, display_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query,
			image.ProductID,
			image.ImageURL,
			image.IsPrimary,
			image.DisplayOrder,
		).Scan(&image.ID, &image.CreatedAt)
	})
	if err != nil {
		return nil, fmt.Errorf("create product image failed: %w", err)
	}
	return image, nil
}

// GetByID retrieves a product image by its ID
func (r *ProductImageRepoImpl) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (model.ProductImage, error) {
	query := `
		SELECT id, product_id, image_url, is_primary, display_order, created_at
		FROM product_images
		WHERE id = $1
		LIMIT 1
	`
	var img model.ProductImage
	err := r.db.QueryRow(ctx, query, id).Scan(
		&img.ID,
		&img.ProductID,
		&img.ImageURL,
		&img.IsPrimary,
		&img.DisplayOrder,
		&img.CreatedAt,
	)
	if err != nil {
		return model.ProductImage{}, err
	}
	return img, nil
}

// GetByProductID retrieves all images for a specific product
func (r *ProductImageRepoImpl) GetByProductID(
	ctx context.Context,
	productID uuid.UUID,
) ([]model.ProductImage, error) {
	query := `
		SELECT id, product_id, image_url, is_primary, display_order, created_at
		FROM product_images
		WHERE product_id = $1
		ORDER BY display_order ASC
	`
	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []model.ProductImage
	for rows.Next() {
		var img model.ProductImage
		err := rows.Scan(
			&img.ID,
			&img.ProductID,
			&img.ImageURL,
			&img.IsPrimary,
			&img.DisplayOrder,
			&img.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return images, nil
}

// Update updates an existing product image
func (r *ProductImageRepoImpl) Update(
	ctx context.Context,
	image *model.ProductImage,
) (*model.ProductImage, error) {
	query := `
		UPDATE product_images 
		SET image_url = $1,
			is_primary = $2,
			display_order = $3
		WHERE id = $4
		RETURNING product_id, created_at
	`
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query,
			image.ImageURL,
			image.IsPrimary,
			image.DisplayOrder,
			image.ID,
		).Scan(&image.ProductID, &image.CreatedAt)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update product image: %w", err)
	}
	return image, nil
}

func (r *ProductImageRepoImpl) UpdateBulk(
	ctx context.Context,
	images []*model.ProductImage,
) ([]*model.ProductImage, error) {
	if len(images) == 0 {
		return images, nil
	}

	query := `
		UPDATE product_images 
		SET image_url = $1,
			is_primary = $2,
			display_order = $3
		WHERE id = $4
		RETURNING product_id, created_at
	`

	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}

		// Queue semua update
		for _, image := range images {
			batch.Queue(query,
				image.ImageURL,
				image.IsPrimary,
				image.DisplayOrder,
				image.ID,
			)
		}

		br := tx.SendBatch(ctx, batch)
		defer br.Close()

		// Scan hasil satu per satu
		for _, image := range images {
			err := br.QueryRow().Scan(
				&image.ProductID,
				&image.CreatedAt,
			)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to bulk update product images: %w", err)
	}

	return images, nil
}

// Delete removes a product image and reorders remaining images
func (r *ProductImageRepoImpl) Delete(
	ctx context.Context,
	productID uuid.UUID,
	imageID uuid.UUID,
) error {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		// Step 1: Get the display_order of the image to be deleted
		var deletedOrder int
		queryGetOrder := `
			SELECT display_order 
			FROM product_images 
			WHERE id = $1 AND product_id = $2
		`
		err := tx.QueryRow(ctx, queryGetOrder, imageID, productID).Scan(&deletedOrder)
		if err != nil {
			return fmt.Errorf("failed to get image order: %w", err)
		}

		// Step 2: Delete the image
		queryDelete := `
			DELETE FROM product_images 
			WHERE id = $1 AND product_id = $2
		`
		cmdTag, err := tx.Exec(ctx, queryDelete, imageID, productID)
		if err != nil {
			return fmt.Errorf("failed to delete image: %w", err)
		}
		if cmdTag.RowsAffected() == 0 {
			return fmt.Errorf("image with id %s not found for product %s", imageID, productID)
		}

		// Step 3: Reorder remaining images (decrease order by 1 for images with higher order)
		queryReorder := `
			UPDATE product_images 
			SET display_order = display_order - 1 
			WHERE product_id = $1 AND display_order > $2
		`
		_, err = tx.Exec(ctx, queryReorder, productID, deletedOrder)
		if err != nil {
			return fmt.Errorf("failed to reorder images: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("delete product image failed: %w", err)
	}
	return nil
}

func (r *ProductImageRepoImpl) CreateBulk(
	ctx context.Context,
	images []*model.ProductImage,
) ([]*model.ProductImage, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO product_images 
				(product_id, image_url, is_primary, display_order)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at
		`

		batch := &pgx.Batch{}

		for _, image := range images {
			batch.Queue(
				query,
				image.ProductID,
				image.ImageURL,
				image.IsPrimary,
				image.DisplayOrder,
			)
		}

		// Kirim batch ke database
		br := tx.SendBatch(ctx, batch)
		defer br.Close()

		for i := range images {
			err := br.QueryRow().Scan(
				&images[i].ID,
				&images[i].CreatedAt,
			)
			if err != nil {
				return fmt.Errorf("failed to insert image: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("create bulk product images failed: %w", err)
	}

	return images, nil
}

func (r *ProductImageRepoImpl) DeleteBulk(
	ctx context.Context,
	productID uuid.UUID,
	imageIDs []uuid.UUID,
) error {

	if len(imageIDs) == 0 {
		return nil
	}

	return pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {

		// 1️⃣ Delete
		deleteQuery := `
			DELETE FROM product_images
			WHERE product_id = $1
			AND id = ANY($2)
		`

		_, err := tx.Exec(ctx, deleteQuery, productID, imageIDs)
		if err != nil {
			return fmt.Errorf("delete images failed: %w", err)
		}

		// 2️⃣ Reorder + update primary
		reorderQuery := `
			WITH reordered AS (
				SELECT id,
					   ROW_NUMBER() OVER (ORDER BY display_order ASC) - 1 AS new_order
				FROM product_images
				WHERE product_id = $1
			)
			UPDATE product_images p
			SET display_order = r.new_order,
			    is_primary = (r.new_order = 0)
			FROM reordered r
			WHERE p.id = r.id;
		`

		_, err = tx.Exec(ctx, reorderQuery, productID)
		if err != nil {
			return fmt.Errorf("reorder failed: %w", err)
		}

		return nil
	})
}

func (r *ProductImageRepoImpl) GetAllByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]model.ProductImage, error) {

	if len(ids) == 0 {
		return []model.ProductImage{}, nil
	}

	query := `
		SELECT id, product_id, image_url, is_primary, display_order, created_at
		FROM product_images
		WHERE id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []model.ProductImage

	for rows.Next() {
		var img model.ProductImage

		err := rows.Scan(
			&img.ID,
			&img.ProductID,
			&img.ImageURL,
			&img.IsPrimary,
			&img.DisplayOrder,
			&img.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		images = append(images, img)
	}

	return images, nil
}
