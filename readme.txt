Сервис получает заказы через NATS Streaming, сохраняет их в PostgreSQL и отдаёт по HTTP.

Запуск:

(в папке проекта)
psql -U postgres -d orders -f init_db.sql

Пароль: 1234

docker run -d --name nats-streaming -p 4222:4222 -p 8222:8222 nats-streaming:latest

(в папке проекта)
go mod tidy
go run .

