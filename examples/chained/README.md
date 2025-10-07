# Chained Resolvers Example

This example demonstrates using multiple tenant resolution strategies with fallback.

## How It Works

The middleware tries resolvers in order:

1. **Header First**: `X-Tenant-ID: tenant1`
2. **Subdomain Second**: `tenant2.localhost`
3. **Query Parameter Third**: `?tenant=tenant3`

The first resolver that successfully extracts a tenant wins!

## Running

```bash
go run main.go
```

## Testing Different Resolution Methods

### Using Header (Priority 1)

```bash
curl -X POST http://localhost:3000/products \
  -H "X-Tenant-ID: company1" \
  -H "Content-Type: application/json" \
  -d '{"name": "Laptop", "price": 999.99}'

curl http://localhost:3000/products \
  -H "X-Tenant-ID: company1"
```

### Using Subdomain (Priority 2)

```bash
curl -X POST http://company2.localhost:3000/products \
  -H "Content-Type: application/json" \
  -d '{"name": "Mouse", "price": 29.99}'

curl http://company2.localhost:3000/products
```

### Using Query Parameter (Priority 3)

```bash
curl -X POST "http://localhost:3000/products?tenant=company3" \
  -H "Content-Type: application/json" \
  -d '{"name": "Keyboard", "price": 79.99}'

curl "http://localhost:3000/products?tenant=company3"
```

## Use Cases

This pattern is useful when:

- **API clients** send tenant via header
- **Web users** access via subdomain
- **Testing/debugging** with query parameters
- **Mobile apps** with custom resolution logic
