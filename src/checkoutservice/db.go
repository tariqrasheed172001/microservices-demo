package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

type OrderRecord struct {
	OrderID             string
	UserID              string
	Email               string
	CurrencyCode        string
	TotalUnits          int64
	TotalNanos          int32
	PaymentTransaction  string
	ShippingTrackingID  string
	ShippingAddressJSON []byte
	ItemsJSON           []byte
	CreatedAt           time.Time
}

func NewOrderRepository(ctx context.Context) (*OrderRepository, error) {
	dsn := os.Getenv("ORDER_DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("environment variable ORDER_DB_DSN not set")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}
	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	repo := &OrderRepository{pool: pool}
	if err := repo.ensureSchema(ctx); err != nil {
		repo.pool.Close()
		return nil, err
	}
	return repo, nil
}

func (r *OrderRepository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

func (r *OrderRepository) ensureSchema(ctx context.Context) error {
	// Create a simple orders table using JSONB for address and items
	const schema = `
	CREATE TABLE IF NOT EXISTS orders (
	  order_id TEXT PRIMARY KEY,
	  user_id TEXT NOT NULL,
	  email TEXT NOT NULL,
	  currency_code TEXT NOT NULL,
	  total_units BIGINT NOT NULL,
	  total_nanos INT NOT NULL,
	  payment_transaction TEXT NOT NULL,
	  shipping_tracking_id TEXT NOT NULL,
	  shipping_address JSONB NOT NULL,
	  items JSONB NOT NULL,
	  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
	CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at DESC);
	`
	_, err := r.pool.Exec(ctx, schema)
	return err
}

func (r *OrderRepository) SaveOrder(ctx context.Context, rec OrderRecord) error {
	const insertSQL = `
	INSERT INTO orders (
	  order_id, user_id, email, currency_code, total_units, total_nanos,
	  payment_transaction, shipping_tracking_id, shipping_address, items, created_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	ON CONFLICT (order_id) DO NOTHING;
	`
	_, err := r.pool.Exec(ctx, insertSQL,
		rec.OrderID,
		rec.UserID,
		rec.Email,
		rec.CurrencyCode,
		rec.TotalUnits,
		rec.TotalNanos,
		rec.PaymentTransaction,
		rec.ShippingTrackingID,
		rec.ShippingAddressJSON,
		rec.ItemsJSON,
		rec.CreatedAt,
	)
	return err
}

func marshalToJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}


