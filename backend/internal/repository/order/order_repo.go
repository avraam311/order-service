package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avraam311/order-service/backend/internal/models"
)

var (
	ErrTxBegin           = errors.New("failed to begin transaction")
	ErrTxCommit          = errors.New("failed to commit transaction")
	ErrInsertOrder       = errors.New("failed to insert into orders")
	ErrInsertDelivery    = errors.New("failed to insert into delivery")
	ErrInsertPayment     = errors.New("failed to insert into payment")
	ErrInsertItem        = errors.New("failed to insert into items")
	ErrOrderNotFound     = errors.New("order not found")
	ErrScanRow           = errors.New("failed to scan row")
	ErrGetItemsByOrderId = errors.New("failed to get items by order ID")
	ErrItemScanFailed    = errors.New("failed to scan order items")
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) SaveOrder(ctx context.Context, order *models.Order) (uuid.UUID, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return uuid.Nil, fmt.Errorf("backend/internal/repository/order_repo.go: %w", ErrTxBegin)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
			return
		}

		if commitErr := tx.Commit(ctx); commitErr != nil {
			tx.Commit(ctx)
			err = fmt.Errorf("backend/internal/repository/order_repo.go: %w", ErrTxCommit)
		}
	}()

	orderQuery := `
	INSERT INTO orders (
		order_uid, track_number, entry, locale, internal_signature, customer_id,
		delivery_service, shardkey, sm_id, oof_shard
	) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING order_uid;
	`

	err = tx.QueryRow(ctx, orderQuery, order.OrderID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerId, order.DeliveryService, order.Shardkey, order.SmId, order.OofShard,
	).Scan(&order.OrderID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("backend/internal/repository/order_repo.go: %w", ErrInsertOrder)
	}

	d := order.Delivery
	deliveryQuery := `
	INSERT INTO delivery (
	    order_uid, name, phone, zip, city, address, region, email
	) VALUES($1, $2, $3, $4, $5, $6, $7, $8);
	`

	_, err = tx.Exec(ctx, deliveryQuery,
		order.OrderID, d.Name, d.Phone, d.Zip, d.City, d.Address, d.Region, d.Email)
	if err != nil {
		return uuid.Nil, fmt.Errorf("backend/internal/repository/order_repo.go: %w", ErrInsertDelivery)
	}

	p := order.Payment
	paymentQuery := `
		INSERT INTO payment (
			order_uid, transaction, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
	`
	_, err = tx.Exec(ctx, paymentQuery,
		order.OrderID, p.Transaction, p.RequestID, p.Currency, p.Provider,
		p.Amount, p.PaymentDT, p.Bank, p.DeliveryCost, p.GoodsTotal, p.CustomFee)
	if err != nil {
		return uuid.Nil, fmt.Errorf("backend/internal/repository/order_repo.go: %w", ErrInsertPayment)
	}

	itemQuery := `
		INSERT INTO items (
			order_id, chrt_id, track_number, price, rid,
			name, sale, size, total_price, nm_id, brand, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
	`
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, itemQuery,
			order.OrderID, item.ChrtID, item.TrackNumber, item.Price, item.RID,
			item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return uuid.Nil, fmt.Errorf("backend/internal/repository/order_repo.go: %w", ErrInsertItem)
		}
	}

	return order.OrderID, nil
}

func (r *Repository) GetOrderById(ctx context.Context, orderID uuid.UUID) (*models.Order, error) {
	query := `
	SELECT
		o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id,
		o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
	
		d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
	
		p.transaction, p.request_id, p.currency, p.provider,
		p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
	FROM orders o
	JOIN delivery d ON o.order_uid = d.order_uid
	JOIN payment p ON o.order_uid = p.order_uid
	WHERE o.order_uid = $1;
	`

	row := r.db.QueryRow(ctx, query, orderID)

	var o models.Order
	var d models.Delivery
	var p models.Payment

	err := row.Scan(
		&o.OrderID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature, &o.CustomerId,
		&o.DeliveryService, &o.Shardkey, &o.SmId, &o.DateCreated, &o.OofShard,

		&d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email,

		&p.Transaction, &p.RequestID, &p.Currency, &p.Provider,
		&p.Amount, &p.PaymentDT, &p.Bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("backend/internal/repository/order_repo.go, get order by id: %w", ErrOrderNotFound)
		}

		return nil, fmt.Errorf("backend/internal/repository/order_repo.go, scan row: %w", ErrScanRow)
	}

	o.Delivery = d
	o.Payment = p

	return &o, err
}

func (r *Repository) GetItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.Item, error) {
	query := `
	SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
	FROM items
	WHERE order_id = $1;
	`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("backend/internal/repository/order_repo.go, get items by order id: %w", ErrGetItemsByOrderId)
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		err = rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale,
			&item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("backend/internal/repository/order_repo.go, scan item row: %w", ErrItemScanFailed)
		}

		items = append(items, item)
	}

	return items, nil
}
