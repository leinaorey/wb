// db.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB


type Order struct {
	ID                int64     `json:"id"`
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	DeliveryName      string    `json:"delivery_name"`
	DeliveryPhone     string    `json:"delivery_phone"`
	DeliveryZip       string    `json:"delivery_zip"`
	DeliveryCity      string    `json:"delivery_city"`
	DeliveryAddress   string    `json:"delivery_address"`
	DeliveryRegion    string    `json:"delivery_region"`
	DeliveryEmail     string    `json:"delivery_email"`
	PaymentTransaction string   `json:"payment_transaction"`
	PaymentRequestID  string    `json:"payment_request_id"`
	PaymentCurrency   string    `json:"payment_currency"`
	PaymentProvider   string    `json:"payment_provider"`
	PaymentAmount     int       `json:"payment_amount"`
	PaymentDt         time.Time `json:"payment_dt"`
	PaymentBank       string    `json:"payment_bank"`
	DeliveryCost      int       `json:"delivery_cost"`
	GoodsTotal        int       `json:"goods_total"`
	CustomFee         int       `json:"custom_fee"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	Items []Item `json:"items"`
}

// Item — структура для работы с таблицей order_items
type Item struct {
	ID           int64     `json:"id"`
	OrderUID     string    `json:"order_uid"`
	ChrtID       int64     `json:"chrt_id"`
	TrackNumber  string    `json:"track_number"`
	Price        int       `json:"price"`
	Rid          string    `json:"rid"`
	Name         string    `json:"name"`
	Sale         int       `json:"sale"`
	Size         string    `json:"size"`
	TotalPrice   int       `json:"total_price"`
	NmID         int64     `json:"nm_id"`
	Brand        string    `json:"brand"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func initDB() {
	connStr := "postgresql://postgres:1234@localhost:5432/orders?sslmode=disable"
	var err error
	DB, err = sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Не удалось пингануть БД:", err)
	}

	log.Println("✅ Подключение к PostgreSQL установлено")
}

func saveToDB(order Order) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Вставка или обновление заказа в таблице orders
	_, err = tx.Exec(`
		INSERT INTO orders (
			order_uid, track_number, entry,
			delivery_name, delivery_phone, delivery_zip, delivery_city,
			delivery_address, delivery_region, delivery_email,
			payment_transaction, payment_request_id, payment_currency,
			payment_provider, payment_amount, payment_dt, payment_bank,
			delivery_cost, goods_total, custom_fee,
			locale, internal_signature, customer_id, delivery_service,
			shardkey, sm_id, date_created, oof_shard
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28)
		ON CONFLICT (order_uid) DO UPDATE SET
			track_number = EXCLUDED.track_number,
			entry = EXCLUDED.entry,
			delivery_name = EXCLUDED.delivery_name,
			delivery_phone = EXCLUDED.delivery_phone,
			delivery_zip = EXCLUDED.delivery_zip,
			delivery_city = EXCLUDED.delivery_city,
			delivery_address = EXCLUDED.delivery_address,
			delivery_region = EXCLUDED.delivery_region,
			delivery_email = EXCLUDED.delivery_email,
			payment_transaction = EXCLUDED.payment_transaction,
			payment_request_id = EXCLUDED.payment_request_id,
			payment_currency = EXCLUDED.payment_currency,
			payment_provider = EXCLUDED.payment_provider,
			payment_amount = EXCLUDED.payment_amount,
			payment_dt = EXCLUDED.payment_dt,
			payment_bank = EXCLUDED.payment_bank,
			delivery_cost = EXCLUDED.delivery_cost,
			goods_total = EXCLUDED.goods_total,
			custom_fee = EXCLUDED.custom_fee,
			locale = EXCLUDED.locale,
			internal_signature = EXCLUDED.internal_signature,
			customer_id = EXCLUDED.customer_id,
			delivery_service = EXCLUDED.delivery_service,
			shardkey = EXCLUDED.shardkey,
			sm_id = EXCLUDED.sm_id,
			date_created = EXCLUDED.date_created,
			oof_shard = EXCLUDED.oof_shard,
			updated_at = NOW()
	`,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.DeliveryName,
		order.DeliveryPhone,
		order.DeliveryZip,
		order.DeliveryCity,
		order.DeliveryAddress,
		order.DeliveryRegion,
		order.DeliveryEmail,
		order.PaymentTransaction,
		order.PaymentRequestID,
		order.PaymentCurrency,
		order.PaymentProvider,
		order.PaymentAmount,
		order.PaymentDt,
		order.PaymentBank,
		order.DeliveryCost,
		order.GoodsTotal,
		order.CustomFee,
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
		return err
	}

	// 2. Удаляем старые позиции (на случай обновления)
	_, err = tx.Exec("DELETE FROM order_items WHERE order_uid = $1", order.OrderUID)
	if err != nil {
		return err
	}

	// 3.  Обязательно: устанавливаем order_uid для каждой позиции
	for i := range order.Items {
		order.Items[i].OrderUID = order.OrderUID
	}

	// 4. Вставляем новые позиции
	for _, item := range order.Items {
		if item.OrderUID == "" {
			return fmt.Errorf("пустой order_uid в позиции")
		}
		_, err = tx.Exec(`
			INSERT INTO order_items (
				order_uid, chrt_id, track_number, price, rid, name,
				sale, size, total_price, nm_id, brand, status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`,
			item.OrderUID,
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
			return err
		}
	}

	return tx.Commit()
}

func restoreCacheFromDB() {
	rows, err := DB.Query(`
		SELECT 
			id, order_uid, track_number, entry,
			delivery_name, delivery_phone, delivery_zip, delivery_city,
			delivery_address, delivery_region, delivery_email,
			payment_transaction, payment_request_id, payment_currency,
			payment_provider, payment_amount, payment_dt, payment_bank,
			delivery_cost, goods_total, custom_fee,
			locale, internal_signature, customer_id, delivery_service,
			shardkey, sm_id, date_created, oof_shard,
			created_at, updated_at
		FROM orders
	`)
	if err != nil {
		log.Fatal("Ошибка восстановления кэша из БД:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.ID,
			&order.OrderUID,
			&order.TrackNumber,
			&order.Entry,
			&order.DeliveryName,
			&order.DeliveryPhone,
			&order.DeliveryZip,
			&order.DeliveryCity,
			&order.DeliveryAddress,
			&order.DeliveryRegion,
			&order.DeliveryEmail,
			&order.PaymentTransaction,
			&order.PaymentRequestID,
			&order.PaymentCurrency,
			&order.PaymentProvider,
			&order.PaymentAmount,
			&order.PaymentDt,
			&order.PaymentBank,
			&order.DeliveryCost,
			&order.GoodsTotal,
			&order.CustomFee,
			&order.Locale,
			&order.InternalSignature,
			&order.CustomerID,
			&order.DeliveryService,
			&order.Shardkey,
			&order.SmID,
			&order.DateCreated,
			&order.OofShard,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			log.Printf("❌ Ошибка сканирования строки заказа: %v", err)
			continue
		}

		// Загружаем связанные позиции
		itemsRows, err := DB.Query(`
			SELECT 
				id, order_uid, chrt_id, track_number, price, rid, name,
				sale, size, total_price, nm_id, brand, status, created_at
			FROM order_items 
			WHERE order_uid = $1
		`, order.OrderUID)
		if err != nil {
			log.Printf("❌ Ошибка получения позиций для заказа %s: %v", order.OrderUID, err)
			continue
		}

		for itemsRows.Next() {
			var item Item
			err := itemsRows.Scan(
				&item.ID,
				&item.OrderUID,
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
				&item.CreatedAt,
			)
			if err != nil {
				log.Printf("❌ Ошибка сканирования позиции: %v", err)
				continue
			}
			order.Items = append(order.Items, item)
		}
		itemsRows.Close()

		cacheMutex.Lock()
		cache[order.OrderUID] = order
		cacheMutex.Unlock()

		log.Printf("🔄 Восстановлено из БД: %s", order.OrderUID)
	}
}