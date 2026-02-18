package payment

import (
	"backEnd-RingoTechLife/internal/common/model"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepositoryInterface interface {
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*model.Payment, error)
	SubmitProof(ctx context.Context, orderID uuid.UUID, proofImageURL string) error
	Approve(ctx context.Context, paymentID uuid.UUID, adminID uuid.UUID, note *string) error
	Reject(ctx context.Context, paymentID uuid.UUID, adminID uuid.UUID, note string) error
	GetPendingPayments(ctx context.Context) ([]model.Payment, error)
}

type PaymentRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepositoryImpl {
	return &PaymentRepositoryImpl{
		db: pool,
	}
}

func (p *PaymentRepositoryImpl) GetByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) (*model.Payment, error) {
	query := `
		SELECT id, order_id, status, amount, proof_image, admin_note, verified_by,
		       created_at, updated_at, submitted_at, verified_at
		FROM payments
		WHERE order_id = $1
		LIMIT 1
	`
	var payment model.Payment
	err := p.db.QueryRow(ctx, query, orderID).Scan(
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
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (p *PaymentRepositoryImpl) SubmitProof(
	ctx context.Context,
	orderID uuid.UUID,
	proofImageURL string,
) error {
	return pgx.BeginFunc(ctx, p.db, func(tx pgx.Tx) error {
		// 1. Update payment
		paymentQuery := `
			UPDATE payments
			SET proof_image = $1,
			    status = $2,
			    submitted_at = NOW(),
			    updated_at = NOW()
			WHERE order_id = $3
		`
		res, err := tx.Exec(ctx, paymentQuery,
			proofImageURL,
			model.PaymentStatusSubmitted,
			orderID,
		)
		if err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("payment not found for order")
		}

		// 2. Update order status
		orderQuery := `
			UPDATE orders
			SET status = $1, updated_at = NOW()
			WHERE id = $2
		`
		_, err = tx.Exec(ctx, orderQuery, model.OrderStatusWaitingConfirmation, orderID)
		if err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}

		return nil
	})
}

func (p *PaymentRepositoryImpl) Approve(
	ctx context.Context,
	paymentID uuid.UUID,
	adminID uuid.UUID,
	note *string,
) error {
	return pgx.BeginFunc(ctx, p.db, func(tx pgx.Tx) error {
		// 1. Get order_id from payment
		var orderID uuid.UUID
		err := tx.QueryRow(ctx, `SELECT order_id FROM payments WHERE id = $1`, paymentID).Scan(&orderID)
		if err != nil {
			return fmt.Errorf("failed to get order_id: %w", err)
		}

		// 2. Update payment
		paymentQuery := `
			UPDATE payments
			SET status = $1,
			    verified_by = $2,
			    admin_note = $3,
			    verified_at = NOW(),
			    updated_at = NOW()
			WHERE id = $4
		`
		res, err := tx.Exec(ctx, paymentQuery,
			model.PaymentStatusApproved,
			adminID,
			note,
			paymentID,
		)
		if err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("payment not found")
		}

		// 3. Update order status
		orderQuery := `
			UPDATE orders
			SET status = $1, confirmed_at = NOW(), updated_at = NOW()
			WHERE id = $2
		`
		_, err = tx.Exec(ctx, orderQuery, model.OrderStatusConfirmed, orderID)
		if err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}

		return nil
	})
}

func (p *PaymentRepositoryImpl) Reject(
	ctx context.Context,
	paymentID uuid.UUID,
	adminID uuid.UUID,
	note string,
) error {
	return pgx.BeginFunc(ctx, p.db, func(tx pgx.Tx) error {
		// 1. Get order_id from payment
		var orderID uuid.UUID
		err := tx.QueryRow(ctx, `SELECT order_id FROM payments WHERE id = $1`, paymentID).Scan(&orderID)
		if err != nil {
			return fmt.Errorf("failed to get order_id: %w", err)
		}

		// 2. Update payment
		paymentQuery := `
			UPDATE payments
			SET status = $1,
			    verified_by = $2,
			    admin_note = $3,
			    verified_at = NOW(),
			    updated_at = NOW()
			WHERE id = $4
		`
		res, err := tx.Exec(ctx, paymentQuery,
			model.PaymentStatusRejected,
			adminID,
			note,
			paymentID,
		)
		if err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("payment not found")
		}
		itemsQuery := `
			SELECT product_id, quantity
			FROM order_items
			WHERE order_id = $1
		`
		rows, err := tx.Query(ctx, itemsQuery, orderID)
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

		// Restore stock
		stockQuery := `UPDATE products SET stock = stock + $1 WHERE id = $2`
		for _, restore := range restores {
			_, err := tx.Exec(ctx, stockQuery, restore.Quantity, restore.ProductID)
			if err != nil {
				return fmt.Errorf("failed to restore stock: %w", err)
			}
		}

		// Update order status
		orderQuery := `
			UPDATE orders
			SET status = $1, cancelled_at = NOW(), updated_at = NOW()
			WHERE id = $2
		`
		_, err = tx.Exec(ctx, orderQuery, model.OrderStatusCancelled, orderID)
		if err != nil {
			return fmt.Errorf("failed to cancel order: %w", err)
		}

		return nil
	})
}

func (p *PaymentRepositoryImpl) GetPendingPayments(ctx context.Context) ([]model.Payment, error) {
	query := `
		SELECT id, order_id, status, amount, proof_image, admin_note, verified_by,
		       created_at, updated_at, submitted_at, verified_at
		FROM payments
		WHERE status = $1
		ORDER BY submitted_at ASC
	`
	rows, err := p.db.Query(ctx, query, model.PaymentStatusSubmitted)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []model.Payment
	for rows.Next() {
		var payment model.Payment
		err := rows.Scan(
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
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, payment)
	}

	return payments, nil
}
