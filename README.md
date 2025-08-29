# Тестове завдання Software Engineer (Golang)

## Getting started

To start the project run

```
docker compose up -d && docker compose logs -f
```

Check notifications

```
docker logs notifications -f
```

### Create product.

\*Price is stored in cents

```
curl -X POST "http://localhost:8081/products" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Product",
    "description": "A test product",
    "price": 100
  }'
```

Example response

```json
{
  "data": {
    "id": "87d9fb79-680b-4390-9c2f-dd2423040fe1",
    "name": "Test product",
    "description": "test description",
    "price": 100,
    "created_at": "2025-08-29T10:47:10.709142Z"
  },
  "success": true
}
```

### Delete Product

```
curl -X DELETE "http://localhost:8081/products/:uuid"
```

Example response

```json
{
  "data": {
    "id": "87d9fb79-680b-4390-9c2f-dd2423040fe1",
    "name": "Test product",
    "description": "test description",
    "price": 100,
    "created_at": "2025-08-29T10:47:10.709142Z"
  },
  "success": true
}
```

### Get Products

```
curl -X GET "http://localhost:8081/products/?limit=3&page=2"
```

Example Response

```json
{
  "data": [
    {
      "id": "13b1f060-08e2-41fb-b620-12c1f9fc8294",
      "name": "Pdfgh2",
      "description": "jjjjs",
      "price": 100,
      "created_at": "2025-08-29T09:51:00.121263Z"
    },
    {
      "id": "e9a9fa48-8674-49b7-a571-6514578a2864",
      "name": "Pdfgh3",
      "description": "jjjjs",
      "price": 100,
      "created_at": "2025-08-29T09:51:04.95464Z"
    },
    {
      "id": "8171cbdc-d05a-4a8c-b9aa-325b4f14c7b0",
      "name": "Pdfgh4",
      "description": "jjjjs",
      "price": 100,
      "created_at": "2025-08-29T09:51:09.395615Z"
    }
  ],
  "page": 2,
  "pages": 5,
  "size": 3,
  "success": true,
  "total": 15
}
```

### Get Metrics

```
curl -X GET "http://localhost:8081/metrics"
```

---

## Technologies

- Go (Gin framework)
- PostgreSQL (with raw SQL)
- Apache Kafka (event streaming)
- Prometheus (metrics collection)
- golang-migrate (database migrations)

## Some Notes

If we absolutely dont want to lose messages on producer side we can implement Transactional Outbox or CDC (Change Data Capture).
And for consumer side send failed messages to dead letter queue.
