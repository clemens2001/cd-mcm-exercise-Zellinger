# Docker Documentation

## Dockerfile analysis

The project uses a multi-stage Docker build. This keeps the final runtime image smaller because the Go compiler, downloaded modules, source files, and build cache stay in the build stage instead of being copied into the final image.

### Stage 1: Build stage

```dockerfile
FROM golang:1.26-alpine AS builder
```

This stage uses the official Go image based on Alpine Linux. It contains the Go toolchain needed to download dependencies and compile the API.

```dockerfile
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-server ./cmd/api
```

The build first copies only `go.mod` and `go.sum` so Docker can cache dependency downloads. After that, the source code is copied and the API binary is built from `./cmd/api`.

`CGO_ENABLED=0` disables cgo, which means the binary is built without depending on C libraries from the build environment. This is useful for containers because the compiled Go binary can run more reliably in a minimal runtime image.

`GOOS=linux` ensures the binary is built for Linux, which is the operating system used inside the final container.

### Stage 2: Runtime stage

```dockerfile
FROM alpine:3.19
```

The runtime image starts from Alpine Linux instead of the full Go image. This image does not need the Go compiler because the API binary was already built in the previous stage.

```dockerfile
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /api-server .

EXPOSE 8080

ENTRYPOINT ["./api-server"]
```

Only the compiled `/api-server` binary is copied from the builder stage. The container exposes port `8080` and starts the API through the binary.

## Why two stages?

The first stage is optimized for building the application. It needs the Go compiler and project source code.

The second stage is optimized for running the application. It only needs the compiled binary and a minimal Linux environment.

This separation reduces the final image size and avoids shipping build tools that are not needed at runtime.

## Image size comparison

To compare the image sizes, we can build the full multi-stage image, and then build just the first stage (the builder) to simulate a single-stage build. The single-stage version keeps the Go base image and build files, so it is much larger.

### Commands used

Built the optimized multi-stage image:
```bash
docker build -t product-catalog:multi-stage .
```

Built only the first stage (simulating a single-stage build) using the --target flag:

```bash
docker build --target builder -t product-catalog:single-stage .
```

Then compared the sizes:

```bash
docker images | grep product-catalog
```

### Results

Output:

```
cd-mcm-exercise-Zellinger % docker images | grep product-catalog

WARNING: This output is designed for human readability. For machine-readable output, please use --format.
product-catalog:multi-stage                                    937c8aa960e3       28.4MB         8.87MB        
product-catalog:single-stage                                   86578c696d4f        499MB         97.5MB    
```

| Image | Full Size | Notes |
| --- | --- | --- |
| `product-catalog:multi-stage` | 28.4MB | Final image using both stages. |
| `product-catalog:single-stage` | 499MB | Image containing the Go toolchain and build environment (stopped at `builder` stage). |

Short comparison:

The multi-stage image is significantly smaller (28.4MB vs 499MB) because it only contains the compiled Go binary running on a minimal Alpine Linux runtime. The single-stage image still contains the entire Go compiler, downloaded dependency modules, build cache, and source code, which bloats the final image size with tools that aren't actually needed to run the application.

## Docker Compose

The application can be started together with PostgreSQL through Docker Compose:

```bash
docker compose up --build
```

The `db` service runs PostgreSQL with the configured database `productcatalog`. The `api` service receives `DB_HOST=db`, so `cmd/api/main.go` selects the PostgreSQL-backed store. The API waits until the database health check passes before it starts.

The `pgdata` Docker volume stores the PostgreSQL data under `/var/lib/postgresql/data`, so data should survive container restarts.

## CRUD test results

Paste your curl commands and responses here after testing the running Docker Compose setup.

### Create products

```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Laptop","price":999.99}'

curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Mouse","price":25.50}'

curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Keyboard","price":45.00}'
```

### List products

```bash
curl http://localhost:8080/products
```

### Update product

```bash
curl -X PUT http://localhost:8080/products/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Gaming Laptop","price":1299.99}'
```

### Delete product

```bash
curl -X DELETE http://localhost:8080/products/2
```

### Verify deleted product is gone

```bash
curl -i http://localhost:8080/products/2
```


Result:

```
cd-mcm-exercise-Zellinger > curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Laptop","price":999.99}'

{"id":1,"name":"Laptop","price":999.99}

cd-mcm-exercise-Zellinger > curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Mouse","price":25.50}'

{"id":2,"name":"Mouse","price":25.5}

cd-mcm-exercise-Zellinger > curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Keyboard","price":45.00}'

{"id":3,"name":"Keyboard","price":45}

cd-mcm-exercise-Zellinger > curl http://localhost:8080/products

[{"id":1,"name":"Laptop","price":999.99},{"id":2,"name":"Mouse","price":25.5},{"id":3,"name":"Keyboard","price":45}]

cd-mcm-exercise-Zellinger > curl -X PUT http://localhost:8080/products/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Gaming Laptop","price":1299.99}'

{"id":1,"name":"Gaming Laptop","price":1299.99}

cd-mcm-exercise-Zellinger > curl http://localhost:8080/products           
  
[{"id":1,"name":"Gaming Laptop","price":1299.99},{"id":2,"name":"Mouse","price":25.5},{"id":3,"name":"Keyboard","price":45}]

cd-mcm-exercise-Zellinger > curl -X DELETE http://localhost:8080/products/2

{"result":"success"}

cd-mcm-exercise-Zellinger > curl -i http://localhost:8080/products/2

HTTP/1.1 404 Not Found
Content-Type: application/json
Date: Sat, 02 May 2026 13:39:39 GMT
Content-Length: 30

{"error":"Product not found"}

cd-mcm-exercise-Zellinger > curl http://localhost:8080/products            

[{"id":1,"name":"Gaming Laptop","price":1299.99},{"id":3,"name":"Keyboard","price":45}]
```