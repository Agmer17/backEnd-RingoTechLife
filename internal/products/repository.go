package products

import (
	"backEnd-RingoTechLife/internal/common/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrProductNotFound = errors.New("product not found")
var ErrFkTagsConstraint = errors.New("Tag tidak ditemukan!")
var ErrConflictSlugName = errors.New("Slug sudah tersedia di database!")
var ErrNameConflict = errors.New("nama produk sudah terdaftar di database! masukan nama lainnnya")
var ErrConflicSku = errors.New("Sku produk sudah tersedia di database! harap masukan yg lain!")

type ProductRepositoryInterface interface {
	// Basic CRUD
	Create(ctx context.Context, product *model.Product) (*model.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (model.Product, error)
	GetBySlug(ctx context.Context, slug string) (model.Product, error)
	Update(ctx context.Context, product *model.Product) (*model.Product, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Listing & Filtering
	GetAllProducts(ctx context.Context) ([]model.Product, error)
	GetProductsByCategorySlug(ctx context.Context, categorySlug string) ([]model.Product, error)
	GetProductsByStatus(ctx context.Context, status string) ([]model.Product, error)
	GetFeaturedProducts(ctx context.Context) ([]model.Product, error)

	// Validation & Checks
	IsProductExistsById(ctx context.Context, id uuid.UUID) (bool, model.Product, error)
	IsProductExistsBySlug(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error)
	IsProductExistsBySKU(ctx context.Context, sku string, excludeID *uuid.UUID) (bool, error)

	// Stock Management
	UpdateStock(ctx context.Context, id uuid.UUID, quantity int) error

	// Search
	SearchProducts(ctx context.Context, keyword string) ([]model.Product, error)
}

type ProductRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewProductsRepository(pool *pgxpool.Pool) *ProductRepositoryImpl {

	return &ProductRepositoryImpl{
		pool: pool,
	}
}

func (r *ProductRepositoryImpl) Create(
	ctx context.Context,
	product *model.Product,
) (*model.Product, error) {

	query := `
		INSERT INTO products 
			(category_id, name, slug, description, brand, condition, 
			 price, stock, sku, specifications, status, is_featured, weight)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at
	`

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query,
			product.CategoryID,
			product.Name,
			product.Slug,
			product.Description,
			product.Brand,
			product.Condition,
			product.Price,
			product.Stock,
			product.SKU,
			product.Specifications,
			product.Status,
			product.IsFeatured,
			product.Weight,
		).Scan(&product.ID, &product.CreatedAt)
	})

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "products_slug_key":
					return nil, ErrConflictSlugName
				case "products_sku_key":
					return nil, ErrConflicSku
				}
			}

			if pgErr.Code == "23503" {
				return nil, ErrFkTagsConstraint
			}
		}

		return nil, fmt.Errorf("create product failed: %w", err)
	}

	return product, nil
}

// GetByID retrieves a product by its ID
func (r *ProductRepositoryImpl) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (model.Product, error) {
	query := `
		SELECT 
			p.id, p.category_id, p.name, p.slug, p.description, p.brand, 
			p.condition, p.price, p.stock, p.sku, p.specifications, 
			p.status, p.is_featured, p.weight, p.created_at,
			c.id, c.name, c.slug, c.description, c.created_at,
			COALESCE(
				json_agg(
					json_build_object(
						'id', pi.id,
						'product_id', pi.product_id,
						'image_url', pi.image_url,
						'is_primary', pi.is_primary,
						'display_order', pi.display_order,
						'created_at', pi.created_at AT TIME ZONE 'UTC'
					) ORDER BY pi.display_order
				) FILTER (WHERE pi.id IS NOT NULL),
				'[]'::json
			) as images
		FROM products p
		INNER JOIN categories c ON p.category_id = c.id
		LEFT JOIN product_images pi ON p.id = pi.product_id
		WHERE p.id = $1
		GROUP BY p.id, c.id
		LIMIT 1
	`

	var (
		p          model.Product
		cat        model.Category
		imagesJSON []byte
	)

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.CategoryID,
		&p.Name,
		&p.Slug,
		&p.Description,
		&p.Brand,
		&p.Condition,
		&p.Price,
		&p.Stock,
		&p.SKU,
		&p.Specifications,
		&p.Status,
		&p.IsFeatured,
		&p.Weight,
		&p.CreatedAt,
		&cat.ID,
		&cat.Name,
		&cat.Slug,
		&cat.Description,
		&cat.CreatedAt,
		&imagesJSON,
	)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return model.Product{}, ErrProductNotFound
		}

		return model.Product{}, err
	}

	// Unmarshal images JSON to struct
	if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
		return model.Product{}, err
	}

	p.Category = cat

	return p, nil
}

func (r *ProductRepositoryImpl) GetBySlug(
	ctx context.Context,
	slug string,
) (model.Product, error) {
	query := `
		SELECT 
			p.id, p.category_id, p.name, p.slug, p.description, p.brand, 
			p.condition, p.price, p.stock, p.sku, p.specifications, 
			p.status, p.is_featured, p.weight, p.created_at,
			c.id, c.name, c.slug, c.description, c.created_at,
			COALESCE(
				json_agg(
					json_build_object(
						'id', pi.id,
						'product_id', pi.product_id,
						'image_url', pi.image_url,
						'is_primary', pi.is_primary,
						'display_order', pi.display_order,
						'created_at', pi.created_at AT TIME ZONE 'UTC'
					) ORDER BY pi.display_order
				) FILTER (WHERE pi.id IS NOT NULL),
				'[]'::json
			) as images
		FROM products p
		INNER JOIN categories c ON p.category_id = c.id
		LEFT JOIN product_images pi ON p.id = pi.product_id
		WHERE p.slug = $1
		GROUP BY p.id, c.id
		LIMIT 1
	`

	var (
		p          model.Product
		cat        model.Category
		imagesJSON []byte
	)

	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&p.ID,
		&p.CategoryID,
		&p.Name,
		&p.Slug,
		&p.Description,
		&p.Brand,
		&p.Condition,
		&p.Price,
		&p.Stock,
		&p.SKU,
		&p.Specifications,
		&p.Status,
		&p.IsFeatured,
		&p.Weight,
		&p.CreatedAt,
		&cat.ID,
		&cat.Name,
		&cat.Slug,
		&cat.Description,
		&cat.CreatedAt,
		&imagesJSON,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Product{}, ErrProductNotFound
		}

		return model.Product{}, err
	}

	// Unmarshal images JSON to struct
	if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
		return model.Product{}, err
	}

	p.Category = cat

	return p, nil
}

// Update updates an existing product
func (r *ProductRepositoryImpl) Update(
	ctx context.Context,
	product *model.Product,
) (*model.Product, error) {

	query := `
		UPDATE products 
		SET category_id = $1,
		    name = $2,
		    slug = $3,
		    description = $4,
		    brand = $5,
		    condition = $6,
		    price = $7,
		    stock = $8,
		    sku = $9,
		    specifications = $10,
		    status = $11,
		    is_featured = $12,
		    weight = $13
		WHERE id = $14
		RETURNING created_at
	`

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query,
			product.CategoryID,
			product.Name,
			product.Slug,
			product.Description,
			product.Brand,
			product.Condition,
			product.Price,
			product.Stock,
			product.SKU,
			product.Specifications,
			product.Status,
			product.IsFeatured,
			product.Weight,
			product.ID,
		).Scan(&product.CreatedAt)
	})

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "products_slug_key":
					return nil, ErrConflictSlugName
				case "products_sku_key":
					return nil, ErrConflicSku
				}
			}

			if pgErr.Code == "23503" {
				return nil, ErrFkTagsConstraint
			}
		}

		return nil, fmt.Errorf("update product failed: %w", err)
	}

	return product, nil
}

// Delete removes a product by ID
func (r *ProductRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		res, err := tx.Exec(ctx, query, id)
		if err != nil {
			return err
		}

		if res.RowsAffected() == 0 {
			return ErrProductNotFound
		}

		return nil
	})

	return err
}

func (r *ProductRepositoryImpl) GetAllProducts(
	ctx context.Context,
) ([]model.Product, error) {
	query := `
		SELECT 
			p.id, p.category_id, p.name, p.slug, p.description, p.brand, 
			p.condition, p.price, p.stock, p.sku, p.specifications, 
			p.status, p.is_featured, p.weight, p.created_at,
			c.id, c.name, c.slug, c.description, c.created_at,
			COALESCE(
				json_agg(
					json_build_object(
						'id', pi.id,
						'product_id', pi.product_id,
						'image_url', pi.image_url,
						'is_primary', pi.is_primary,
						'display_order', pi.display_order,
						'created_at', pi.created_at AT TIME ZONE 'UTC'
					) ORDER BY pi.display_order
				) FILTER (WHERE pi.id IS NOT NULL),
				'[]'::json
			) as images
		FROM products p
		INNER JOIN categories c ON p.category_id = c.id
		LEFT JOIN product_images pi ON p.id = pi.product_id
		GROUP BY p.id, c.id
		ORDER BY p.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []model.Product{}

	for rows.Next() {
		var (
			p          model.Product
			cat        model.Category
			imagesJSON []byte
		)

		err := rows.Scan(
			&p.ID,
			&p.CategoryID,
			&p.Name,
			&p.Slug,
			&p.Description,
			&p.Brand,
			&p.Condition,
			&p.Price,
			&p.Stock,
			&p.SKU,
			&p.Specifications,
			&p.Status,
			&p.IsFeatured,
			&p.Weight,
			&p.CreatedAt,
			&cat.ID,
			&cat.Name,
			&cat.Slug,
			&cat.Description,
			&cat.CreatedAt,
			&imagesJSON,
		)

		if err != nil {
			return nil, err
		}

		// Unmarshal images JSON to struct
		if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
			return nil, err
		}

		p.Category = cat
		products = append(products, p)
	}

	return products, rows.Err()
}

// GetProductsByCategory retrieves products by category ID
func (r *ProductRepositoryImpl) GetProductsByCategorySlug(
	ctx context.Context,
	categorySlug string,
) ([]model.Product, error) {
	query := `
		SELECT 
			p.id, p.category_id, p.name, p.slug, p.description, p.brand, 
			p.condition, p.price, p.stock, p.sku, p.specifications, 
			p.status, p.is_featured, p.weight, p.created_at,
			c.id, c.name, c.slug, c.description, c.created_at,
			COALESCE(
				json_agg(
					json_build_object(
						'id', pi.id,
						'product_id', pi.product_id,
						'image_url', pi.image_url,
						'is_primary', pi.is_primary,
						'display_order', pi.display_order,
						'created_at', pi.created_at AT TIME ZONE 'UTC'
					) ORDER BY pi.display_order
				) FILTER (WHERE pi.id IS NOT NULL),
				'[]'::json
			) as images
		FROM products p
		INNER JOIN categories c ON p.category_id = c.id
		LEFT JOIN product_images pi ON p.id = pi.product_id
		WHERE c.slug = $1
		GROUP BY p.id, c.id
		ORDER BY p.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, categorySlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []model.Product{}

	for rows.Next() {
		var (
			p          model.Product
			cat        model.Category
			imagesJSON []byte
		)

		err := rows.Scan(
			&p.ID,
			&p.CategoryID,
			&p.Name,
			&p.Slug,
			&p.Description,
			&p.Brand,
			&p.Condition,
			&p.Price,
			&p.Stock,
			&p.SKU,
			&p.Specifications,
			&p.Status,
			&p.IsFeatured,
			&p.Weight,
			&p.CreatedAt,
			&cat.ID,
			&cat.Name,
			&cat.Slug,
			&cat.Description,
			&cat.CreatedAt,
			&imagesJSON,
		)

		if err != nil {
			return nil, err
		}

		// Unmarshal images JSON to struct
		if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
			return nil, err
		}

		p.Category = cat
		products = append(products, p)
	}

	return products, rows.Err()
}

func (r *ProductRepositoryImpl) GetProductsByStatus(
	ctx context.Context,
	status string,
) ([]model.Product, error) {
	query := `
		SELECT 
			p.id, p.category_id, p.name, p.slug, p.description, p.brand, 
			p.condition, p.price, p.stock, p.sku, p.specifications, 
			p.status, p.is_featured, p.weight, p.created_at,
			c.id, c.name, c.slug, c.description, c.created_at,
			COALESCE(
				json_agg(
					json_build_object(
						'id', pi.id,
						'product_id', pi.product_id,
						'image_url', pi.image_url,
						'is_primary', pi.is_primary,
						'display_order', pi.display_order,
						'created_at', pi.created_at AT TIME ZONE 'UTC'
					) ORDER BY pi.display_order
				) FILTER (WHERE pi.id IS NOT NULL),
				'[]'::json
			) as images
		FROM products p
		INNER JOIN categories c ON p.category_id = c.id
		LEFT JOIN product_images pi ON p.id = pi.product_id
		WHERE p.status = $1
		GROUP BY p.id, c.id
		ORDER BY p.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []model.Product{}

	for rows.Next() {
		var (
			p          model.Product
			cat        model.Category
			imagesJSON []byte
		)

		err := rows.Scan(
			&p.ID,
			&p.CategoryID,
			&p.Name,
			&p.Slug,
			&p.Description,
			&p.Brand,
			&p.Condition,
			&p.Price,
			&p.Stock,
			&p.SKU,
			&p.Specifications,
			&p.Status,
			&p.IsFeatured,
			&p.Weight,
			&p.CreatedAt,
			&cat.ID,
			&cat.Name,
			&cat.Slug,
			&cat.Description,
			&cat.CreatedAt,
			&imagesJSON,
		)

		if err != nil {
			return nil, err
		}

		// Unmarshal images JSON to struct
		if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
			return nil, err
		}

		p.Category = cat
		products = append(products, p)
	}

	return products, rows.Err()
}

func (r *ProductRepositoryImpl) GetFeaturedProducts(
	ctx context.Context,
) ([]model.Product, error) {
	query := `
		SELECT 
			p.id, p.category_id, p.name, p.slug, p.description, p.brand, 
			p.condition, p.price, p.stock, p.sku, p.specifications, 
			p.status, p.is_featured, p.weight, p.created_at,
			c.id, c.name, c.slug, c.description, c.created_at,
			COALESCE(
				json_agg(
					json_build_object(
						'id', pi.id,
						'product_id', pi.product_id,
						'image_url', pi.image_url,
						'is_primary', pi.is_primary,
						'display_order', pi.display_order,
						'created_at', pi.created_at AT TIME ZONE 'UTC'
					) ORDER BY pi.display_order
				) FILTER (WHERE pi.id IS NOT NULL),
				'[]'::json
			) as images
		FROM products p
		INNER JOIN categories c ON p.category_id = c.id
		LEFT JOIN product_images pi ON p.id = pi.product_id
		WHERE p.is_featured = true
		GROUP BY p.id, c.id
		ORDER BY p.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []model.Product{}

	for rows.Next() {
		var (
			p          model.Product
			cat        model.Category
			imagesJSON []byte
		)

		err := rows.Scan(
			&p.ID,
			&p.CategoryID,
			&p.Name,
			&p.Slug,
			&p.Description,
			&p.Brand,
			&p.Condition,
			&p.Price,
			&p.Stock,
			&p.SKU,
			&p.Specifications,
			&p.Status,
			&p.IsFeatured,
			&p.Weight,
			&p.CreatedAt,
			&cat.ID,
			&cat.Name,
			&cat.Slug,
			&cat.Description,
			&cat.CreatedAt,
			&imagesJSON,
		)

		if err != nil {
			return nil, err
		}

		// Unmarshal images JSON to struct
		if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
			return nil, err
		}

		p.Category = cat
		products = append(products, p)
	}

	return products, rows.Err()
}

// IsProductExistsById checks if a product exists by ID and returns the product
func (r *ProductRepositoryImpl) IsProductExistsById(
	ctx context.Context,
	id uuid.UUID,
) (bool, model.Product, error) {

	query := `
		SELECT id, category_id, name, slug, description, brand, condition,
		       price, stock, sku, specifications, status, is_featured, weight, created_at
		FROM products
		WHERE id = $1
		LIMIT 1
	`

	var p model.Product
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.CategoryID,
		&p.Name,
		&p.Slug,
		&p.Description,
		&p.Brand,
		&p.Condition,
		&p.Price,
		&p.Stock,
		&p.SKU,
		&p.Specifications,
		&p.Status,
		&p.IsFeatured,
		&p.Weight,
		&p.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return false, model.Product{}, nil
		}
		return false, model.Product{}, err
	}

	return true, p, nil
}

// IsProductExistsBySlug checks if a product with the given slug exists
func (r *ProductRepositoryImpl) IsProductExistsBySlug(
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
				FROM products
				WHERE slug = $1
				  AND id != $2
			)
		`
		args = []any{slug, *excludeID}
	} else {
		query = `
			SELECT EXISTS (
				SELECT 1
				FROM products
				WHERE slug = $1
			)
		`
		args = []any{slug}
	}

	var exists bool
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// IsProductExistsBySKU checks if a product with the given SKU exists
func (r *ProductRepositoryImpl) IsProductExistsBySKU(
	ctx context.Context,
	sku string,
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
				FROM products
				WHERE sku = $1
				  AND id != $2
			)
		`
		args = []any{sku, *excludeID}
	} else {
		query = `
			SELECT EXISTS (
				SELECT 1
				FROM products
				WHERE sku = $1
			)
		`
		args = []any{sku}
	}

	var exists bool
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// UpdateStock updates the stock quantity of a product
func (r *ProductRepositoryImpl) UpdateStock(
	ctx context.Context,
	id uuid.UUID,
	quantity int,
) error {

	query := `
		UPDATE products 
		SET stock = $1
		WHERE id = $2
	`

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		res, err := tx.Exec(ctx, query, quantity, id)
		if err != nil {
			return err
		}

		if res.RowsAffected() == 0 {
			return fmt.Errorf("product with id %s not found", id)
		}

		return nil
	})

	return err
}

// SearchProducts searches products by name using full-text search (Indonesian)
func (r *ProductRepositoryImpl) SearchProducts(
	ctx context.Context,
	keyword string,
) ([]model.Product, error) {
	query := `
        SELECT 
            p.id, p.category_id, p.name, p.slug, p.description, p.brand, 
            p.condition, p.price, p.stock, p.sku, p.specifications, 
            p.status, p.is_featured, p.weight, p.created_at,
            c.id, c.name, c.slug, c.description, c.created_at,
            COALESCE(
                json_agg(
                    json_build_object(
                        'id', pi.id,
                        'product_id', pi.product_id,
                        'image_url', pi.image_url,
                        'is_primary', pi.is_primary,
                        'display_order', pi.display_order,
                        'created_at', pi.created_at AT TIME ZONE 'UTC'
                    ) ORDER BY pi.display_order
                ) FILTER (WHERE pi.id IS NOT NULL),
                '[]'::json
            ) as images
        FROM products p
        INNER JOIN categories c ON p.category_id = c.id
        LEFT JOIN product_images pi ON p.id = pi.product_id
        WHERE to_tsvector('indonesian', p.name) @@ plainto_tsquery('indonesian', $1)
        GROUP BY p.id, c.id
        ORDER BY p.created_at DESC
    `

	rows, err := r.pool.Query(ctx, query, keyword)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []model.Product{}

	for rows.Next() {
		var (
			p          model.Product
			cat        model.Category
			imagesJSON []byte
		)

		err := rows.Scan(
			&p.ID,
			&p.CategoryID,
			&p.Name,
			&p.Slug,
			&p.Description,
			&p.Brand,
			&p.Condition,
			&p.Price,
			&p.Stock,
			&p.SKU,
			&p.Specifications,
			&p.Status,
			&p.IsFeatured,
			&p.Weight,
			&p.CreatedAt,
			&cat.ID,
			&cat.Name,
			&cat.Slug,
			&cat.Description,
			&cat.CreatedAt,
			&imagesJSON,
		)

		if err != nil {
			return nil, err
		}

		// Unmarshal images JSON to struct
		if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
			return nil, err
		}

		p.Category = cat
		products = append(products, p)
	}

	return products, rows.Err()
}
