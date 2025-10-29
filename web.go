// web.go
package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Обработчик главной страницы с формой поиска
func listOrdersHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем результат поиска (если есть)
	query := r.URL.Query().Get("q")
	var foundOrder *Order = nil

	if query != "" {
		cacheMutex.RLock()
		if order, exists := cache[query]; exists {
			foundOrder = &order
		}
		cacheMutex.RUnlock()
	}

	// Формируем список всех заказов для отображения
	cacheMutex.RLock()
	orders := make([]Order, 0, len(cache))
	for _, order := range cache {
		orders = append(orders, order)
	}
	cacheMutex.RUnlock()

	data := struct {
		Orders      []Order
		SearchQuery string
		FoundOrder  *Order
	}{
		Orders:      orders,
		SearchQuery: query,
		FoundOrder:  foundOrder,
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Заказы</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; }
		table { border-collapse: collapse; width: 100%; margin-top: 15px; }
		th, td { border: 1px solid #ddd; padding: 10px; text-align: left; }
		th { background-color: #f2f2f2; }
		a { color: #0066cc; text-decoration: none; }
		a:hover { text-decoration: underline; }
		form { margin-bottom: 20px; }
		input[type="text"] { padding: 6px; width: 300px; }
		input[type="submit"] { padding: 6px 12px; }
		.result { margin-top: 20px; padding: 15px; background-color: #e9f7ef; border: 1px solid #27ae60; }
	</style>
</head>
<body>
	<h1>Список заказов</h1>

	<form method="GET">
		<label for="q">Поиск по Order UID:</label><br>
		<input type="text" id="q" name="q" value="{{.SearchQuery}}" placeholder="Введите order_uid">
		<input type="submit" value="Найти">
	</form>

	{{if .FoundOrder}}
	<div class="result">
		<h2>Найден заказ:</h2>
		<p><strong>Order UID:</strong> {{.FoundOrder.OrderUID}}</p>
		<p><strong>Трек:</strong> {{.FoundOrder.TrackNumber}}</p>
		<p><strong>Клиент:</strong> {{.FoundOrder.DeliveryName}} ({{.FoundOrder.DeliveryCity}})</p>
		<p><strong>Дата:</strong> {{.FoundOrder.DateCreated.Format "2006-01-02 15:04:05"}}</p>
		<p><strong>Сумма:</strong> {{.FoundOrder.PaymentAmount}}</p>
		<a href="/order/{{.FoundOrder.OrderUID}}">Просмотреть детали</a>
	</div>
	{{end}}

	<h2>Все заказы</h2>
	<table>
		<tr>
			<th>Order UID</th>
			<th>Трек</th>
			<th>Клиент</th>
			<th>Город</th>
			<th>Дата</th>
			<th>Сумма</th>
			<th>Действие</th>
		</tr>
		{{range .Orders}}
		<tr>
			<td>{{.OrderUID}}</td>
			<td>{{.TrackNumber}}</td>
			<td>{{.DeliveryName}}</td>
			<td>{{.DeliveryCity}}</td>
			<td>{{.DateCreated.Format "2006-01-02 15:04"}}</td>
			<td>{{.PaymentAmount}}</td>
			<td><a href="/order/{{.OrderUID}}">Просмотр</a></td>
		</tr>
		{{end}}
	</table>
</body>
</html>
`

	t, _ := template.New("list").Parse(tmpl)
	t.Execute(w, data)
}

// Обработчик детальной страницы заказа
func orderDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	cacheMutex.RLock()
	order, exists := cache[uid]
	cacheMutex.RUnlock()

	if !exists {
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Заказ {{.OrderUID}}</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; }
		h1 { color: #333; }
		.info { margin: 10px 0; }
		table { border-collapse: collapse; width: 100%; margin-top: 15px; }
		th, td { border: 1px solid #ddd; padding: 10px; text-align: left; }
		th { background-color: #f2f2f2; }
		a { color: #0066cc; }
	</style>
</head>
<body>
	<h1>Заказ {{.OrderUID}}</h1>
	<div class="info"><strong>Трек:</strong> {{.TrackNumber}}</div>
	<div class="info"><strong>Клиент:</strong> {{.DeliveryName}} ({{.DeliveryCity}})</div>
	<div class="info"><strong>Дата:</strong> {{.DateCreated.Format "2006-01-02 15:04:05"}}</div>
	<div class="info"><strong>Сумма:</strong> {{.PaymentAmount}}</div>

	<h2>Товары</h2>
	<table>
		<tr><th>Название</th><th>Бренд</th><th>Цена</th><th>Итого</th><th>Статус</th></tr>
		{{range .Items}}
		<tr>
			<td>{{.Name}}</td>
			<td>{{.Brand}}</td>
			<td>{{.Price}}</td>
			<td>{{.TotalPrice}}</td>
			<td>{{.Status}}</td>
		</tr>
		{{end}}
	</table>
	<br>
	<a href="/">← Назад к списку</a>
</body>
</html>
`

	t, _ := template.New("detail").Parse(tmpl)
	t.Execute(w, order)
}

// JSON API
func apiOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	cacheMutex.RLock()
	order, exists := cache[uid]
	cacheMutex.RUnlock()

	if !exists {
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// Запуск веб-сервера
func StartWebServer() {
	r := mux.NewRouter()
	r.HandleFunc("/", listOrdersHandler).Methods("GET")
	r.HandleFunc("/order/{uid}", orderDetailHandler).Methods("GET")
	r.HandleFunc("/api/order/{uid}", apiOrderHandler).Methods("GET")

	log.Println("Веб-интерфейс доступен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}