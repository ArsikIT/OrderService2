# Order Service

`order-service` owns order data, exposes an HTTP API for end users, calls `payment-service` over gRPC for payment authorization, and exposes a gRPC server-side streaming endpoint for order status updates.

## Architecture

Project layout:

```text
order-service/
|- cmd/order-service/main.go
|- cmd/order-stream-client/main.go
|- internal/domain
|- internal/usecase
|- internal/repository/postgres
|- internal/transport/http
|- internal/transport/grpc
|- internal/app
|- migrations
|- docker-compose.yml
`- README.md
```

Dependency direction:

```text
HTTP handlers -> use cases -> repository/payment ports
gRPC handlers -> use cases -> repository/payment ports
                                   |-> postgres repository
                                   |-> gRPC payment client
                                   `-> order updates publisher
```

Key decisions:

- `domain` contains only business entities, statuses, and business errors.
- `usecase` contains order creation, payment orchestration, cancellation rules, and publishes order status updates after successful DB changes.
- `repository/postgres` contains persistence only.
- `transport/http` stays thin and maps REST requests/responses.
- `transport/grpc` contains server-side streaming for order updates.
- `cmd/order-service/main.go` is the composition root with manual dependency injection.

## Business Rules

- Money uses `int64` cents only.
- Order amount must be greater than `0`.
- New order starts as `Pending`.
- If payment is authorized, order becomes `Paid`.
- If payment is declined, order becomes `Failed`.
- If payment service is unavailable, order becomes `Failed` and API returns `503 Service Unavailable`.
- Only `Pending` orders can be cancelled.
- `Paid` orders cannot be cancelled.

## Database Per Service

This service has its own PostgreSQL container and its own database:

- container: `order-db`
- database: `order_service`
- port: `55433`

`order-service` does not read or write `payment-service` tables. Payment authorization happens only through gRPC.

## Environment Variables

Default values are listed in [.env.example](/C:/Users/hp/order-service/.env.example).

- `HTTP_ADDRESS` default: `:8080`
- `GRPC_ADDRESS` default: `:50052`
- `DATABASE_URL` default: `postgres://postgres:postgres@127.0.0.1:55433/order_service?sslmode=disable`
- `PAYMENT_GRPC_ADDR` default: `127.0.0.1:50051`
- `ORDER_GRPC_ADDR` default for demo client: `127.0.0.1:50052`

## Run

1. Start the order database:

```bash
docker compose up -d
```

2. Make sure `payment-service` is already running with:

- HTTP on `:8081`
- gRPC on `:50051`

3. Run `order-service`:

```bash
go run ./cmd/order-service
```

4. Optional: run the streaming demo client:

```bash
go run ./cmd/order-stream-client <order-id>
```

## HTTP API Examples

Create order:

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: order-123" \
  -d "{\"customer_id\":\"cust-1\",\"item_name\":\"Headphones\",\"amount\":15000}"
```

Get order:

```bash
curl http://localhost:8080/orders/{id}
```

Cancel order:

```bash
curl -X PATCH http://localhost:8080/orders/{id}/cancel
```

## gRPC Streaming Demo

Subscribe to order updates:

```bash
go run ./cmd/order-stream-client <order-id>
```

The stream sends:

- the current order status from the database immediately after subscription
- later status changes published by application logic
- status changes detected from the database by periodic re-checking

Example DB update for defense/demo:

```bash
docker exec -it order-db psql -U postgres -d order_service -c "UPDATE orders SET status = 'Failed' WHERE id = '<order-id>';"
```

## Architecture Diagram

```
REST
```

![img.png](img.png)

```
gRPC
```
![grpc3.png](grpc3.png)