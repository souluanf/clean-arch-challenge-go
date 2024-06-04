FROM golang:1.22.3-alpine

WORKDIR /app

COPY .docker/wait-for-it.sh /app/wait-for-it.sh
COPY .docker/Dockerfile /app/Dockerfile

COPY . .

RUN go install github.com/ktr0731/evans@latest && \
    go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest && \
    go mod tidy

RUN chmod +x wait-for-it.sh

CMD ["sh", "-c", "./wait-for-it.sh mysql && migrate -path=internal/infra/database/migrations -database 'mysql://root:root@tcp(mysql:3306)/orders' -verbose up && go run cmd/ordersystem/wire_gen.go cmd/ordersystem/main.go"]