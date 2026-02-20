package payment

import (
	"backEnd-RingoTechLife/internal/common/model"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type paymentValidationData struct {
	Amount      float64
	IssuerId    uuid.UUID
	OrderStatus string
}

var ErrNoPaymentfound = errors.New("pembayarab dengan id ini tidak ditemukan!")

type PaymentRepositoryInterface interface {
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*model.Payment, error)
	SubmitProof(ctx context.Context, tmp *model.Payment) error
	Approve(ctx context.Context, paymentID uuid.UUID, adminID uuid.UUID, note *string) error
	Reject(ctx context.Context, paymentID uuid.UUID, adminID uuid.UUID, note string) error
	GetPendingPayments(ctx context.Context) ([]model.Payment, error)

	ExistByOrderId(ctx context.Context, orderId uuid.UUID) (bool, error)
	GetPaymentValidationData(ctx context.Context, orderId uuid.UUID) (paymentValidationData, error)
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoPaymentfound
		}
		return nil, err
	}
	return &payment, nil
}

func (p *PaymentRepositoryImpl) SubmitProof(
	ctx context.Context,
	tempData *model.Payment,
) error {
	return pgx.BeginFunc(ctx, p.db, func(tx pgx.Tx) error {
		paymentQuery := `
			UPDATE payments
			SET proof_image = $1,
			    status = $2,
				amount = $3,
			    submitted_at = NOW(),
			    updated_at = NOW()
			WHERE order_id = $4
			RETURNING id,status, submitted_at, updated_at
		`
		err := tx.QueryRow(ctx, paymentQuery,
			tempData.ProofImage,
			model.PaymentStatusSubmitted,
			tempData.Amount,
			tempData.OrderID,
		).Scan(&tempData.ID, &tempData.Status, &tempData.SubmittedAt, &tempData.UpdatedAt)

		if err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

		orderQuery := `
			UPDATE orders
			SET status = $1, updated_at = NOW()
			WHERE id = $2
		`
		_, err = tx.Exec(ctx, orderQuery, model.OrderStatusWaitingConfirmation, tempData.OrderID)
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
		// Update payment + order sekaligus dalam 1 CTE query
		query := `
			WITH payment_update AS (
				UPDATE payments
				SET status = $1,
					verified_by = $2,
					admin_note = $3,
					verified_at = NOW(),
					updated_at = NOW()
				WHERE id = $4
				RETURNING order_id
			)
			UPDATE orders
			SET status = $5, confirmed_at = NOW(), updated_at = NOW()
			FROM payment_update
			WHERE orders.id = payment_update.order_id
			RETURNING orders.id
		`

		var orderID uuid.UUID
		err := tx.QueryRow(ctx, query,
			model.PaymentStatusApproved,
			adminID,
			note,
			paymentID,
			model.OrderStatusConfirmed,
		).Scan(&orderID)

		if err != nil {
			if err == pgx.ErrNoRows {
				return fmt.Errorf("payment not found")
			}
			return fmt.Errorf("failed to approve payment: %w", err)
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
		// 1. Update payment + order + restore stock sekaligus dalam 1 CTE query
		query := `
			WITH payment_update AS (
				UPDATE payments
				SET status = $1,
					verified_by = $2,
					admin_note = $3,
					verified_at = NOW(),
					updated_at = NOW()
				WHERE id = $4
				RETURNING order_id
			),
			order_update AS (
				UPDATE orders
				SET status = $5, cancelled_at = NOW(), updated_at = NOW()
				FROM payment_update
				WHERE orders.id = payment_update.order_id
				RETURNING orders.id
			),
			stock_restore AS (
				UPDATE products
				SET stock = stock + oi.quantity
				FROM order_items oi
				INNER JOIN order_update ON oi.order_id = order_update.id
				WHERE products.id = oi.product_id
				RETURNING products.id
			)
			SELECT order_update.id FROM order_update
		`

		var orderID uuid.UUID
		err := tx.QueryRow(ctx, query,
			model.PaymentStatusRejected,
			adminID,
			note,
			paymentID,
			model.OrderStatusCancelled,
		).Scan(&orderID)

		if err != nil {
			if err == pgx.ErrNoRows {
				return fmt.Errorf("payment not found")
			}
			return fmt.Errorf("failed to reject payment: %w", err)
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

func (p *PaymentRepositoryImpl) ExistByOrderId(ctx context.Context, orderId uuid.UUID) (bool, error) {
	query := `select exist(
		select 1 from payments 
		where order_id = $1
	)`
	var result bool

	err := p.db.QueryRow(ctx, query, orderId).Scan(&result)

	if err != nil {
		return false, err
	}
	return result, nil
}

func (p *PaymentRepositoryImpl) GetPaymentValidationData(ctx context.Context, orderId uuid.UUID) (paymentValidationData, error) {

	query := `
	select o.subtotal, o.user_id, o.status from orders o where o.id = $1
	`

	var tempData paymentValidationData
	err := p.db.QueryRow(ctx, query, orderId).Scan(&tempData.Amount, &tempData.IssuerId, &tempData.OrderStatus)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return paymentValidationData{}, ErrNoPaymentfound
		}
		return paymentValidationData{}, err
	}

	return tempData, nil

}
