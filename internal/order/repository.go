// internal/repository/order/order_repository.go

package order

import (
	"backEnd-RingoTechLife/internal/common/model"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepositoryInterface interface {
	Create(ctx context.Context, order *model.Order, items []model.OrderItem) (*model.Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*model.Order, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Order, error)
	GetByUserIDWithDetails(ctx context.Context, userID uuid.UUID) ([]model.Order, error)
	GetAll(ctx context.Context) ([]model.Order, error)
	GetByStatus(ctx context.Context, status model.OrderStatus) ([]model.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error
	Cancel(ctx context.Context, id uuid.UUID) error
}

type OrderRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepositoryImpl {
	return &OrderRepositoryImpl{
		db: pool,
	}
}

func (r *OrderRepositoryImpl) Create(
	ctx context.Context,
	order *model.Order,
	items []model.OrderItem,
) (*model.Order, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		// 1. Insert order
		orderQuery := `
			INSERT INTO orders (user_id, status, subtotal, total_amount, notes)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at, updated_at
		`
		err := tx.QueryRow(ctx, orderQuery,
			order.UserID,
			order.Status,
			order.Subtotal,
			order.TotalAmount,
			order.Notes,
		).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert order: %w", err)
		}

		// 2. Insert order items + update stock
		itemQuery := `
			INSERT INTO order_items 
				(order_id, product_id, product_name, product_sku, price_at_purchase, quantity, subtotal)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at
		`
		stockQuery := `
			UPDATE products 
			SET stock = stock - $1
			WHERE id = $2 AND stock >= $1
			RETURNING stock
		`

		for i := range items {
			items[i].OrderID = order.ID

			// Insert item
			err := tx.QueryRow(ctx, itemQuery,
				items[i].OrderID,
				items[i].ProductID,
				items[i].ProductName,
				items[i].ProductSKU,
				items[i].PriceAtPurchase,
				items[i].Quantity,
				items[i].Subtotal,
			).Scan(&items[i].ID, &items[i].CreatedAt)
			if err != nil {
				return fmt.Errorf("failed to insert order item: %w", err)
			}

			// Update stock
			var newStock int
			err = tx.QueryRow(ctx, stockQuery, items[i].Quantity, items[i].ProductID).Scan(&newStock)
			if err != nil {
				if err == pgx.ErrNoRows {
					return fmt.Errorf("insufficient stock for product %s", items[i].ProductID)
				}
				return fmt.Errorf("failed to update stock: %w", err)
			}
		}

		// 3. Insert payment record
		paymentQuery := `
			INSERT INTO payments (order_id, status, amount)
			VALUES ($1, $2, $3)
			RETURNING id, created_at, updated_at
		`
		var paymentID uuid.UUID
		var paymentCreatedAt, paymentUpdatedAt interface{}
		err = tx.QueryRow(ctx, paymentQuery,
			order.ID,
			model.PaymentStatusUnpaid,
			order.TotalAmount,
		).Scan(&paymentID, &paymentCreatedAt, &paymentUpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert payment: %w", err)
		}

		order.Items = items
		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *OrderRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	query := `
		SELECT id, user_id, status, subtotal, total_amount, notes,
		       created_at, updated_at, confirmed_at, cancelled_at
		FROM orders
		WHERE id = $1
	`
	var order model.Order
	err := r.db.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.Subtotal,
		&order.TotalAmount,
		&order.Notes,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.ConfirmedAt,
		&order.CancelledAt,
	)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepositoryImpl) GetByIDWithDetails(
	ctx context.Context,
	id uuid.UUID,
) (*model.Order, error) {
	// Get order basic info
	order, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get order items
	itemsQuery := `
		SELECT id, order_id, product_id, product_name, product_sku,
		       price_at_purchase, quantity, subtotal, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
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

	// Get payment info
	paymentQuery := `
		SELECT id, order_id, status, amount, proof_image, admin_note, verified_by,
		       created_at, updated_at, submitted_at, verified_at
		FROM payments
		WHERE order_id = $1
		LIMIT 1
	`
	var payment model.Payment
	err = r.db.QueryRow(ctx, paymentQuery, id).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Status,
		&payment.Amount,
		&payment.ProofImage,
		&payment.AdminNote,
		&payment.VerifiedBy,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.SubmittedAt,
		&payment.VerifiedAt,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	order.Items = items
	if err != pgx.ErrNoRows {
		order.Payment = &payment
	}

	return order, nil
}

func (r *OrderRepositoryImpl) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]model.Order, error) {
	query := `
		SELECT id, user_id, status, subtotal, total_amount, notes,
		       created_at, updated_at, confirmed_at, cancelled_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.Subtotal,
			&order.TotalAmount,
			&order.Notes,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.ConfirmedAt,
			&order.CancelledAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *OrderRepositoryImpl) GetByUserIDWithDetails(
	ctx context.Context,
	userID uuid.UUID,
) ([]model.Order, error) {

	query := `
	SELECT
		o.id,
		o.user_id,
		o.status,
		o.subtotal,
		o.total_amount,
		o.notes,
		o.created_at,
		o.updated_at,
		o.confirmed_at,
		o.cancelled_at,

		COALESCE(
			json_agg(
				jsonb_build_object(
					'id', oi.id,
					'order_id', oi.order_id,
					'product_id', oi.product_id,
					'product_name', oi.product_name,
					'product_sku', oi.product_sku,
					'price_at_purchase', oi.price_at_purchase,
					'quantity', oi.quantity,
					'subtotal', oi.subtotal,
					'created_at', oi.created_at AT TIME ZONE 'UTC'
				)
			) FILTER (WHERE oi.id IS NOT NULL),
			'[]'
		) AS items,

		to_jsonb(p) AS payment

	FROM orders o
	LEFT JOIN order_items oi ON oi.order_id = o.id
	LEFT JOIN payments p ON p.order_id = o.id
	WHERE o.user_id = $1
	GROUP BY o.id, p.id
	ORDER BY o.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order

	for rows.Next() {
		var order model.Order
		var itemsJSON []byte
		var paymentJSON []byte

		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.Subtotal,
			&order.TotalAmount,
			&order.Notes,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.ConfirmedAt,
			&order.CancelledAt,
			&itemsJSON,
			&paymentJSON,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal items
		if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
			return nil, fmt.Errorf("failed to unmarshal items: %w", err)
		}

		// Unmarshal payment (nullable)
		if paymentJSON != nil {
			var payment model.Payment
			if err := json.Unmarshal(paymentJSON, &payment); err == nil {
				order.Payment = &payment
			}
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (r *OrderRepositoryImpl) GetAll(ctx context.Context) ([]model.Order, error) {
	query := `
		SELECT id, user_id, status, subtotal, total_amount, notes,
		       created_at, updated_at, confirmed_at, cancelled_at
		FROM orders
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.Subtotal,
			&order.TotalAmount,
			&order.Notes,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.ConfirmedAt,
			&order.CancelledAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *OrderRepositoryImpl) GetByStatus(
	ctx context.Context,
	status model.OrderStatus,
) ([]model.Order, error) {
	query := `
		SELECT id, user_id, status, subtotal, total_amount, notes,
		       created_at, updated_at, confirmed_at, cancelled_at
		FROM orders
		WHERE status = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.Subtotal,
			&order.TotalAmount,
			&order.Notes,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.ConfirmedAt,
			&order.CancelledAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *OrderRepositoryImpl) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status model.OrderStatus,
) error {
	return pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var query string
		var args []interface{}

		switch status {
		case model.OrderStatusConfirmed:
			query = `
				UPDATE orders
				SET status = $1, confirmed_at = NOW(), updated_at = NOW()
				WHERE id = $2
			`
			args = []interface{}{status, id}
		case model.OrderStatusCancelled:
			query = `
				UPDATE orders
				SET status = $1, cancelled_at = NOW(), updated_at = NOW()
				WHERE id = $2
			`
			args = []interface{}{status, id}
		default:
			query = `
				UPDATE orders
				SET status = $1, updated_at = NOW()
				WHERE id = $2
			`
			args = []interface{}{status, id}
		}

		res, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("order not found")
		}
		return nil
	})
}

func (r *OrderRepositoryImpl) Cancel(ctx context.Context, id uuid.UUID) error {
	return pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		itemsQuery := `
			SELECT product_id, quantity
			FROM order_items
			WHERE order_id = $1
		`
		rows, err := tx.Query(ctx, itemsQuery, id)
		if err != nil {
			return fmt.Errorf("failed to get order items: %w", err)
		}
		defer rows.Close()

		type stockRestore struct {
			ProductID uuid.UUID
			Quantity  int
		}
		var restores []stockRestore

		for rows.Next() {
			var restore stockRestore
			err := rows.Scan(&restore.ProductID, &restore.Quantity)
			if err != nil {
				return fmt.Errorf("failed to scan item: %w", err)
			}
			restores = append(restores, restore)
		}

		// 2. Restore stock for each product
		stockQuery := `
			UPDATE products
			SET stock = stock + $1
			WHERE id = $2
		`
		for _, restore := range restores {
			_, err := tx.Exec(ctx, stockQuery, restore.Quantity, restore.ProductID)
			if err != nil {
				return fmt.Errorf("failed to restore stock: %w", err)
			}
		}

		// 3. Update order status
		orderQuery := `
			UPDATE orders
			SET status = $1, cancelled_at = NOW(), updated_at = NOW()
			WHERE id = $2
		`
		res, err := tx.Exec(ctx, orderQuery, model.OrderStatusCancelled, id)
		if err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("order not found")
		}

		// 4. Update payment status if exists
		paymentQuery := `
			UPDATE payments
			SET status = 'rejected', updated_at = NOW()
			WHERE order_id = $1
		`
		_, err = tx.Exec(ctx, paymentQuery, id)
		if err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

		return nil
	})
}
