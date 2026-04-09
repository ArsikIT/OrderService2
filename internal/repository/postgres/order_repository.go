package postgres

import (
	"context"
	"database/sql"
	"errors"

	"order-service/internal/domain"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, item_name, amount, status, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		order.ID,
		order.CustomerID,
		order.ItemName,
		order.Amount,
		order.Status,
		order.IdempotencyKey,
		order.CreatedAt,
	)
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, item_name, amount, status, idempotency_key, created_at
		FROM orders
		WHERE id = $1
	`
	return r.scanOne(ctx, query, id)
}

func (r *OrderRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, item_name, amount, status, idempotency_key, created_at
		FROM orders
		WHERE idempotency_key = $1
	`
	return r.scanOne(ctx, query, key)
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE orders SET status = $1 WHERE id = $2`, status, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

func (r *OrderRepository) scanOne(ctx context.Context, query string, arg string) (*domain.Order, error) {
	var order domain.Order
	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&order.ID,
		&order.CustomerID,
		&order.ItemName,
		&order.Amount,
		&order.Status,
		&order.IdempotencyKey,
		&order.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}
