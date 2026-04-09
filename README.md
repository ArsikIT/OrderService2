# Order Service

`order-service` is the order bounded context of the assignment platform. It owns only order data and communicates with `payment-service` over REST to authorize payment.

## Architecture

Project layout:

```text
order-service/
|- cmd/order-service/main.go
|- internal/domain
|- internal/usecase
|- internal/repository/postgres
|- internal/transport/http
|- internal/app
|- migrations
|- docker-compose.yml
`- README.md
```

Dependency direction:

```text
HTTP handlers -> use cases -> repository/payment ports
                                  |-> postgres repository
                                  `-> REST payment client
```

Key decisions:

- `domain` contains only business entities, statuses, and business errors.
- `usecase` contains order creation, payment orchestration, and cancellation rules.
- `repository/postgres` contains persistence only.
- `transport/http` stays thin and only maps request/response.
- `cmd/order-service/main.go` is the composition root with manual dependency injection.

## Business Rules

- Money uses `int64` cents only.
- Order amount must be greater than `0`.
- New order starts as `Pending`.
- If payment is authorized, order becomes `Paid`.
- If payment is declined, order becomes `Failed`.
- If payment service is unavailable or times out, order becomes `Failed` and API returns `503 Service Unavailable`.
- Only `Pending` orders can be cancelled.
- `Paid` orders cannot be cancelled.

## Failure Handling

`order-service` uses a shared custom `http.Client` with a `2s` timeout for calls to `payment-service`.

Why the order becomes `Failed` when `payment-service` is unavailable:

- the synchronous payment attempt for this request did not succeed;
- returning `Pending` would leave the order in a misleading intermediate state;
- `Failed` is easier to justify during defense because the final state reflects the failed dependency call.

## Database Per Service

This service has its own PostgreSQL container and its own database:

- container: `order-db`
- database: `order_service`
- port: `55433`

`order-service` must not read or write `payment-service` tables. Communication with `payment-service` happens only through HTTP.

## Run

1. Start the order database:

```bash
docker compose up -d
```

2. Make sure `payment-service` is already running on `http://127.0.0.1:8081`.

3. Run the service:

```bash
go run ./cmd/order-service
```

Default environment values are listed in [.env.example](C:/Users/hp/order-service/.env.example).

## Environment Variables

- `HTTP_ADDRESS` default: `:8080`
- `DATABASE_URL` default: `postgres://postgres:postgres@127.0.0.1:55433/order_service?sslmode=disable`
- `PAYMENT_BASE_URL` default: `http://127.0.0.1:8081`

## API Examples

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

## Architecture Diagram

```mermaid
flowchart LR
    Client --> OrderAPI[Order Service HTTP API]
    OrderAPI --> OrderUC[Order Use Case]
    OrderUC --> OrderRepo[(Order Service DB)]
    OrderUC --> PaymentPort[Payment REST Client]
    PaymentPort --> PaymentAPI[Payment Service /payments]
    
    
```

![img.png](img.png)
