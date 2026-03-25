package servicerequest

import (
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/common/model"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var notFoundError = errors.New("Data tidak ditemukan!")

type ServiceRequestRepositoryInterface interface {
	Create(ctx context.Context, req *model.ServiceRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ServiceRequest, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.ServiceRequest, error)
	GetAll(ctx context.Context) ([]*model.ServiceRequest, error) // untuk admin

	AdminQuote(ctx context.Context, id uuid.UUID, dto *dto.AdminQuoteServiceRequestDTO, adminID uuid.UUID) error
	AdminReject(ctx context.Context, id uuid.UUID, dto *dto.AdminRejectServiceRequestDTO, adminID uuid.UUID) error
	UserAccept(ctx context.Context, id uuid.UUID, orderID uuid.UUID) error
	UserReject(ctx context.Context, id uuid.UUID) error
	Cancel(ctx context.Context, id uuid.UUID) error
}

type ServiceRequestRepository struct {
	pool *pgxpool.Pool
}

func NewServiceRequestRepository(pool *pgxpool.Pool) *ServiceRequestRepository {
	return &ServiceRequestRepository{pool: pool}
}

// ─── READ ────────────────────────────────────────────────────────────────────

func (r *ServiceRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ServiceRequest, error) {
	query := `
		SELECT id, user_id, device_type, device_brand, device_model,
		       problem_description, photo_1, photo_2, photo_3, status,
		       quoted_price, estimated_duration, admin_note, quoted_by,
		       order_id, created_at, updated_at, quoted_at, decided_at
		FROM service_requests
		WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, id)
	sr, err := scanServiceRequest(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, notFoundError
		}
		return nil, err
	}
	return sr, nil
}

func (r *ServiceRequestRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.ServiceRequest, error) {
	query := `
		SELECT id, user_id, device_type, device_brand, device_model,
		       problem_description, photo_1, photo_2, photo_3, status,
		       quoted_price, estimated_duration, admin_note, quoted_by,
		       order_id, created_at, updated_at, quoted_at, decided_at
		FROM service_requests
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("GetByUserID: %w", err)
	}
	defer rows.Close()

	return collectServiceRequests(rows)
}

func (r *ServiceRequestRepository) GetAll(ctx context.Context) ([]*model.ServiceRequest, error) {
	query := `
		SELECT id, user_id, device_type, device_brand, device_model,
		       problem_description, photo_1, photo_2, photo_3, status,
		       quoted_price, estimated_duration, admin_note, quoted_by,
		       order_id, created_at, updated_at, quoted_at, decided_at
		FROM service_requests
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetAll: %w", err)
	}
	defer rows.Close()

	return collectServiceRequests(rows)
}

// ─── CREATE ──────────────────────────────────────────────────────────────────

func (r *ServiceRequestRepository) Create(ctx context.Context, req *model.ServiceRequest) error {
	query := `
		INSERT INTO service_requests (
			id, user_id, device_type, device_brand, device_model,
			problem_description, photo_1, photo_2, photo_3, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)`

	req.ID = uuid.New()
	_, err := r.pool.Exec(ctx, query,
		req.ID, req.UserID, req.DeviceType, req.DeviceBrand, req.DeviceModel,
		req.ProblemDescription, req.Photo1, req.Photo2, req.Photo3, model.StatusPendingReview,
	)
	if err != nil {
		return fmt.Errorf("Create: %w", err)
	}
	return nil
}

// ─── STATUS TRANSITIONS (semua pakai TX) ─────────────────────────────────────

func (r *ServiceRequestRepository) AdminQuote(ctx context.Context, id uuid.UUID, d *dto.AdminQuoteServiceRequestDTO, adminID uuid.UUID) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		query := `
			UPDATE service_requests SET
				status             = 'quoted',
				quoted_price       = $1,
				estimated_duration = $2,
				admin_note         = $3,
				quoted_by          = $4,
				quoted_at          = $5,
				updated_at         = $5
			WHERE id = $6 AND status = 'pending_review'`

		now := time.Now()
		tag, err := tx.Exec(ctx, query,
			d.QuotedPrice, d.EstimatedDuration, d.AdminNote, adminID, now, id,
		)
		if err != nil {
			return fmt.Errorf("AdminQuote: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("AdminQuote: request not found or not in pending_review status")
		}
		return nil
	})
}

func (r *ServiceRequestRepository) AdminReject(ctx context.Context, id uuid.UUID, d *dto.AdminRejectServiceRequestDTO, adminID uuid.UUID) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		query := `
			UPDATE service_requests SET
				status     = 'rejected_by_admin',
				admin_note = $1,
				quoted_by  = $2,
				updated_at = $3
			WHERE id = $4 AND status = 'pending_review'`

		now := time.Now()
		tag, err := tx.Exec(ctx, query, d.AdminNote, adminID, now, id)
		if err != nil {
			return fmt.Errorf("AdminReject: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("AdminReject: request not found or not in pending_review status")
		}
		return nil
	})
}

func (r *ServiceRequestRepository) UserAccept(ctx context.Context, id uuid.UUID, orderID uuid.UUID) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		query := `
			UPDATE service_requests SET
				status     = 'accepted',
				order_id   = $1,
				decided_at = $2,
				updated_at = $2
			WHERE id = $3 AND status = 'quoted'`

		now := time.Now()
		tag, err := tx.Exec(ctx, query, orderID, now, id)
		if err != nil {
			return fmt.Errorf("UserAccept: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("UserAccept: request not found or not in quoted status")
		}
		return nil
	})
}

func (r *ServiceRequestRepository) UserReject(ctx context.Context, id uuid.UUID) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		query := `
			UPDATE service_requests SET
				status     = 'rejected_by_user',
				decided_at = $1,
				updated_at = $1
			WHERE id = $2 AND status = 'quoted'`

		now := time.Now()
		tag, err := tx.Exec(ctx, query, now, id)
		if err != nil {
			return fmt.Errorf("UserReject: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("UserReject: request not found or not in quoted status")
		}
		return nil
	})
}

func (r *ServiceRequestRepository) Cancel(ctx context.Context, id uuid.UUID) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		query := `
			UPDATE service_requests SET
				status     = 'cancelled',
				updated_at = $1
			WHERE id = $2 AND status = 'pending_review'`

		now := time.Now()
		tag, err := tx.Exec(ctx, query, now, id)
		if err != nil {
			return fmt.Errorf("Cancel: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("Cancel: request not found or not in pending_review status")
		}
		return nil
	})
}

// ─── HELPERS ─────────────────────────────────────────────────────────────────

type scannable interface {
	Scan(dest ...any) error
}

func scanServiceRequest(row scannable) (*model.ServiceRequest, error) {
	var sr model.ServiceRequest
	err := row.Scan(
		&sr.ID, &sr.UserID, &sr.DeviceType, &sr.DeviceBrand, &sr.DeviceModel,
		&sr.ProblemDescription, &sr.Photo1, &sr.Photo2, &sr.Photo3, &sr.Status,
		&sr.QuotedPrice, &sr.EstimatedDuration, &sr.AdminNote, &sr.QuotedBy,
		&sr.OrderID, &sr.CreatedAt, &sr.UpdatedAt, &sr.QuotedAt, &sr.DecidedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sr, nil
}

func collectServiceRequests(rows pgx.Rows) ([]*model.ServiceRequest, error) {
	var results []*model.ServiceRequest
	for rows.Next() {
		sr, err := scanServiceRequest(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, sr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
