package repository

import (
	"database/sql"
	"fmt"
	"log"
	"wb-orders-service/models"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) Close() {
	r.db.Close()
}

// SaveOrder сохраняет заказ в БД (транзакционно)
func (r *PostgresRepository) SaveOrder(order *models.Order) error {
	// Сначала проверяем, существует ли уже заказ с таким order_uid
	exists, err := r.orderExists(order.OrderUID)
	if err != nil {
		return fmt.Errorf("failed to check order existence: %v", err)
	}
	if exists {
		return fmt.Errorf("order with UID %s already exists", order.OrderUID)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Вставляем в таблицу orders
	orderQuery := `INSERT INTO orders (
		order_uid, track_number, entry, locale, internal_signature, 
		customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = tx.Exec(orderQuery,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %v", err)
	}

	// Вставляем в таблицу deliveries
	deliveryQuery := `INSERT INTO deliveries (
		order_uid, name, phone, zip, city, address, region, email
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = tx.Exec(deliveryQuery,
		order.OrderUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("failed to insert delivery: %v", err)
	}

	// Вставляем в таблицу payments
	paymentQuery := `INSERT INTO payments (
		transaction, request_id, currency, provider, amount, 
		payment_dt, bank, delivery_cost, goods_total, custom_fee
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = tx.Exec(paymentQuery,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %v", err)
	}

	// Вставляем все items
	itemQuery := `INSERT INTO items (
		order_uid, chrt_id, track_number, price, rid, name, 
		sale, size, total_price, nm_id, brand, status
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	for _, item := range order.Items {
		_, err = tx.Exec(itemQuery,
			order.OrderUID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)
		if err != nil {
			return fmt.Errorf("failed to insert item: %v", err)
		}
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("Order %s saved successfully", order.OrderUID)
	return nil
}

// orderExists проверяет существует ли заказ с указанным order_uid
func (r *PostgresRepository) orderExists(orderUID string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM orders WHERE order_uid = $1)"
	err := r.db.QueryRow(query, orderUID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetOrderByUID возвращает заказ по его UID
func (r *PostgresRepository) GetOrderByUID(orderUID string) (*models.Order, error) {
	// Получаем основные данные заказа
	orderQuery := `SELECT
		order_uid, track_number, entry, locale, internal_signature,
		customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
	FROM orders WHERE order_uid = $1`

	order := &models.Order{}
	err := r.db.QueryRow(orderQuery, orderUID).Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found: %s", orderUID)
		}
		return nil, fmt.Errorf("failed to get order: %v", err)
	}

	// Получаем данные доставки
	deliveryQuery := `SELECT
		name, phone, zip, city, address, region, email
	FROM deliveries WHERE order_uid = $1`

	delivery := &models.Delivery{}
	err = r.db.QueryRow(deliveryQuery, orderUID).Scan(
		&delivery.Name,
		&delivery.Phone,
		&delivery.Zip,
		&delivery.City,
		&delivery.Address,
		&delivery.Region,
		&delivery.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery: %v", err)
	}
	order.Delivery = *delivery

	// Получаем данные платежа
	paymentQuery := `SELECT
		transaction, request_id, currency, provider, amount,
		payment_dt, bank, delivery_cost, goods_total, custom_fee
	FROM payments WHERE transaction = $1`

	payment := &models.Payment{}
	err = r.db.QueryRow(paymentQuery, orderUID).Scan(
		&payment.Transaction,
		&payment.RequestID,
		&payment.Currency,
		&payment.Provider,
		&payment.Amount,
		&payment.PaymentDt,
		&payment.Bank,
		&payment.DeliveryCost,
		&payment.GoodsTotal,
		&payment.CustomFee,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %v", err)
	}
	order.Payment = *payment

	// Получаем все товары
	itemsQuery := `SELECT
		chrt_id, track_number, price, rid, name, sale, size,
		total_price, nm_id, brand, status
	FROM items WHERE order_uid = $1`

	rows, err := r.db.Query(itemsQuery, orderUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %v", err)
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		item := models.Item{}
		err := rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.Rid,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %v", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating items: %v", err)
	}

	order.Items = items
	return order, nil
}

// GetAllOrders возвращает все заказы (для восстановления кэша)
func (r *PostgresRepository) GetAllOrders() ([]models.Order, error) {
	// Получаем все order_uid
	orderUIDsQuery := `SELECT order_uid FROM orders`
	rows, err := r.db.Query(orderUIDsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get order UIDs: %v", err)
	}
	defer rows.Close()

	var orders []models.Order
	var orderUIDs []string

	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, fmt.Errorf("failed to scan order UID: %v", err)
		}
		orderUIDs = append(orderUIDs, orderUID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order UIDs: %v", err)
	}

	// Для каждого order_uid получаем полный заказ
	for _, orderUID := range orderUIDs {
		order, err := r.GetOrderByUID(orderUID)
		if err != nil {
			log.Printf("Warning: failed to get order %s: %v", orderUID, err)
			continue
		}
		orders = append(orders, *order)
	}

	log.Printf("Loaded %d orders from database", len(orders))
	return orders, nil
}

// InitDB создает таблицы если они не существуют
func (r *PostgresRepository) InitDB() error {
    createTablesSQL := `
    CREATE TABLE IF NOT EXISTS orders (
        order_uid VARCHAR(255) PRIMARY KEY,
        track_number VARCHAR(255),
        entry VARCHAR(50),
        locale VARCHAR(10),
        internal_signature VARCHAR(255),
        customer_id VARCHAR(255),
        delivery_service VARCHAR(100),
        shardkey VARCHAR(50),
        sm_id INTEGER,
        date_created TIMESTAMP WITH TIME ZONE,
        oof_shard VARCHAR(50)
    );

    CREATE TABLE IF NOT EXISTS deliveries (
        id SERIAL PRIMARY KEY,
        order_uid VARCHAR(255) REFERENCES orders(order_uid) ON DELETE CASCADE,
        name VARCHAR(255) NOT NULL,
        phone VARCHAR(50),
        zip VARCHAR(50),
        city VARCHAR(100),
        address TEXT,
        region VARCHAR(100),
        email VARCHAR(255)
    );

    CREATE TABLE IF NOT EXISTS payments (
        transaction VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
        request_id VARCHAR(255),
        currency VARCHAR(10),
        provider VARCHAR(100),
        amount INTEGER,
        payment_dt BIGINT,
        bank VARCHAR(100),
        delivery_cost INTEGER,
        goods_total INTEGER,
        custom_fee INTEGER
    );

    CREATE TABLE IF NOT EXISTS items (
        id SERIAL PRIMARY KEY,
        order_uid VARCHAR(255) REFERENCES orders(order_uid) ON DELETE CASCADE,
        chrt_id BIGINT,
        track_number VARCHAR(255),
        price INTEGER,
        rid VARCHAR(255),
        name VARCHAR(255),
        sale INTEGER,
        size VARCHAR(50),
        total_price INTEGER,
        nm_id BIGINT,
        brand VARCHAR(255),
        status INTEGER
    );
    `

    _, err := r.db.Exec(createTablesSQL)
    if err != nil {
        return fmt.Errorf("failed to create tables: %v", err)
    }

    log.Println("Database tables initialized successfully")
    return nil
}
