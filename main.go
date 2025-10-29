// main.go
package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/nats-io/stan.go"
)


type OrderJSON struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []ItemJSON `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       string    `json:"date_created"` // ISO 8601 строка
	OofShard          string    `json:"oof_shard"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int64  `json:"payment_dt"` // Unix timestamp
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type ItemJSON struct {
	ChrtID      int64  `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int64  `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

// Глобальный кэш (должен быть доступен в web.go)
var cache = make(map[string]Order)
var cacheMutex sync.RWMutex

func main() {
	initDB()
	restoreCacheFromDB()

	sc, err := stan.Connect("test-cluster", "order-service", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		log.Fatal("Не удалось подключиться к NATS Streaming:", err)
	}
	defer sc.Close()

	log.Println("Подключен к NATS Streaming")

	_, err = sc.Subscribe("orders", func(msg *stan.Msg) {
		var msgJSON OrderJSON
		if err := json.Unmarshal(msg.Data, &msgJSON); err != nil {
			log.Printf("Ошибка разбора JSON: %v", err)
			return
		}

		if msgJSON.OrderUID == "" {
			log.Println("order_uid отсутствует в сообщении")
			return
		}

		order := Order{
			OrderUID:          msgJSON.OrderUID,
			TrackNumber:       msgJSON.TrackNumber,
			Entry:             msgJSON.Entry,
			DeliveryName:      msgJSON.Delivery.Name,
			DeliveryPhone:     msgJSON.Delivery.Phone,
			DeliveryZip:       msgJSON.Delivery.Zip,
			DeliveryCity:      msgJSON.Delivery.City,
			DeliveryAddress:   msgJSON.Delivery.Address,
			DeliveryRegion:    msgJSON.Delivery.Region,
			DeliveryEmail:     msgJSON.Delivery.Email,
			PaymentTransaction: msgJSON.Payment.Transaction,
			PaymentRequestID:  msgJSON.Payment.RequestID,
			PaymentCurrency:   msgJSON.Payment.Currency,
			PaymentProvider:   msgJSON.Payment.Provider,
			PaymentAmount:     msgJSON.Payment.Amount,
			PaymentBank:       msgJSON.Payment.Bank,
			DeliveryCost:      msgJSON.Payment.DeliveryCost,
			GoodsTotal:        msgJSON.Payment.GoodsTotal,
			CustomFee:         msgJSON.Payment.CustomFee,
			Locale:            msgJSON.Locale,
			InternalSignature: msgJSON.InternalSignature,
			CustomerID:        msgJSON.CustomerID,
			DeliveryService:   msgJSON.DeliveryService,
			Shardkey:          msgJSON.Shardkey,
			SmID:              msgJSON.SmID,
			OofShard:          msgJSON.OofShard,
		}

		order.PaymentDt = time.Unix(msgJSON.Payment.PaymentDt, 0).UTC()
		if t, err := time.Parse(time.RFC3339, msgJSON.DateCreated); err == nil {
			order.DateCreated = t
		} else {
			order.DateCreated = time.Now()
		}

		for _, itemJSON := range msgJSON.Items {
			order.Items = append(order.Items, Item{
				ChrtID:      itemJSON.ChrtID,
				TrackNumber: itemJSON.TrackNumber,
				Price:       itemJSON.Price,
				Rid:         itemJSON.Rid,
				Name:        itemJSON.Name,
				Sale:        itemJSON.Sale,
				Size:        itemJSON.Size,
				TotalPrice:  itemJSON.TotalPrice,
				NmID:        itemJSON.NmID,
				Brand:       itemJSON.Brand,
				Status:      itemJSON.Status,
			})
		}

		if err := saveToDB(order); err != nil {
			log.Printf("❌ Ошибка записи в БД: %v", err)
			return
		}

		cacheMutex.Lock()
		cache[order.OrderUID] = order
		cacheMutex.Unlock()

		log.Printf("✅ Получено и сохранено: %s", order.OrderUID)
	})
	if err != nil {
		log.Fatal("Ошибка подписки:", err)
	}

	// Запуск веб-интерфейса
	StartWebServer()
}