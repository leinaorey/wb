// publish.go
package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/nats-io/stan.go"
)

// PublishOrder отправляет заказ из файла model.json в канал NATS
func PublishOrder() {
	// Чтение файла model.json
	data, err := os.ReadFile("model.json")
	if err != nil {
		log.Fatal("❌ Не удалось прочитать файл model.json:", err)
	}

	// Проверка валидности JSON
	var order map[string]interface{}
	if err := json.Unmarshal(data, &order); err != nil {
		log.Fatal("❌ Неверный формат JSON в model.json:", err)
	}

	// Подключение к NATS Streaming
	sc, err := stan.Connect("test-cluster", "publisher", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		log.Fatal("❌ Ошибка подключения к NATS:", err)
	}
	defer sc.Close()

	// Отправка сообщения
	if err := sc.Publish("orders", data); err != nil {
		log.Fatal("❌ Ошибка отправки сообщения в NATS:", err)
	}

	log.Println("✅ Сообщение из model.json отправлено в канал 'orders'")
}