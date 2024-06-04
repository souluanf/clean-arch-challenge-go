# Clean Arch Challenge GO

Este desafio consiste na criação do use case de listagem de pedidos (orders), implementando diferentes interfaces de
comunicação: REST, gRPC e GraphQL.

## Pré-requisitos

### Clonar o Repositório

Baixe o repositório e acesse a pasta do desafio:

```bash
git clone https://github.com/souluanf/clean-arch-challenge-go.git
cd clean-arch-challenge-go
```

## Execução

### Somente Docker

- Copiar Variáveis de Ambiente

    ```bash
    cp .env.docker .env
    ```

- Execute a subida dos containers:

   ```bash
   docker-compose up -d
   ```
- Verificar os logs do container:

   ```bash
    docker logs -f app
   ```

### Docker e Local

- Copiar Variáveis de Ambiente

    ```bash
    cp .env.local .env
    ```

1. Instale as dependências:

   ```bash
   go install github.com/ktr0731/evans@latest
   go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   go mod tidy
   ```

2. Execute o MySQL e RabbitMQ:

   ```bash
   docker-compose up -d mysql rabbitmq
   ```

3. Execute as migrações:

   ```bash
   migrate -path=internal/infra/database/migrations -database "mysql://root:root@tcp(localhost:3306)/orders" -verbose up
   ```

4. Execute os servidores:

   ```bash
   go run cmd/ordersystem/wire_gen.go cmd/ordersystem/main.go
   ```

## Testando Endpoints

### API REST

Na pasta `/api` temos os arquivos para testes dos nossos endpoints REST na porta 8080.

- **Criar Order:**
    - Arquivo: `create_order.http`
    - Endpoint: `POST /order`

- **Listar Orders:**
    - Arquivo: `list_orders.http`
    - Endpoint: `GET /order`

Ou utilize o comando `curl`:

- **Criar Order:**
  ```bash
  curl -X POST http://localhost:8000/order -d '{"id": "1f8972d9-9054-4dab-8972-d990549dab54", "Price": 17.15, "Tax": 0.25}'
  ```
- **Listar Orders:**
  ```bash
    curl --location 'http://localhost:8000/order'
  ```

### GraphQL

Em [http://localhost:8080](http://localhost:8080), podemos executar os comandos GraphQL.

- **Criar Order:**
  ```graphql
  mutation createOrder {
    createOrder(input: {id: "1f8972d9-9054-4dab-8972-d990549dab53", Price: 17.15, Tax: 0.25}) {
      id
      Price
      Tax
      FinalPrice
    }
  }
  ```

- **Listar Orders:**
  ```graphql
  query listOrders {
    listOrders {
      id
      Price
      Tax
      FinalPrice
    }
  }
  ```

### gRPC

1. Execute o cliente gRPC:

   ```bash
   evans -r repl -p 50051
   ```

2. Conecte-se ao package `pb`:

   ```bash
   package pb
   ```

3. Conecte-se ao serviço de Orders:

   ```bash
   service OrderService
   ```

4. **Criar Order:**

   ```bash
   call CreateOrder
   ```

5. **Listar Orders:**

   ```bash
   call ListOrders
   ```

### RabbitMQ

- Acesse o RabbitMQ em [http://localhost:15672](http://localhost:15672) com usuário `guest` e senha `guest`.
- Acesse a fila [order_created_queue](http://localhost:15672/#/queues/%2F/order_created_queue)
- Verifique as mensagens enviadas