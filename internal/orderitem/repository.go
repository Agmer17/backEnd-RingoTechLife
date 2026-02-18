package orderitem

import (
	"backEnd-RingoTechLife/internal/common/model"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderItemRepositoryInterface interface {
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]model.OrderItem, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.OrderItem, error)
}

type OrderItemRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewOrderItemRepository(pool *pgxpool.Pool) *OrderItemRepositoryImpl {
	return &OrderItemRepositoryImpl{
		db: pool,
	}
}

func (r *OrderItemRepositoryImpl) GetByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) ([]model.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, product_name, product_sku,
		       price_at_purchase, quantity, subtotal, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var item model.OrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductName,
			&item.ProductSKU,
			&item.PriceAtPurchase,
			&item.Quantity,
			&item.Subtotal,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *OrderItemRepositoryImpl) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, product_name, product_sku,
		       price_at_purchase, quantity, subtotal, created_at
		FROM order_items
		WHERE id = $1
	`
	var item model.OrderItem
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.OrderID,
		&item.ProductID,
		&item.ProductName,
		&item.ProductSKU,
		&item.PriceAtPurchase,
		&item.Quantity,
		&item.Subtotal,
		&item.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}
