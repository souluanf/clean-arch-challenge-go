package main

import (
	"clean-arch-challenge-go/configs"
	"clean-arch-challenge-go/internal/event/handler"
	"clean-arch-challenge-go/internal/infra/graph"
	"clean-arch-challenge-go/internal/infra/grpc/pb"
	"clean-arch-challenge-go/internal/infra/grpc/service"
	"clean-arch-challenge-go/internal/infra/web/webserver"
	"clean-arch-challenge-go/internal/usecase"
	"clean-arch-challenge-go/pkg/events"
	"database/sql"
	"fmt"
	graphqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/go-sql-driver/mysql"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db := initDB(config)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	}(db)

	rabbitMQChannel := getRabbitMQChannel(config)

	eventDispatcher := events.NewEventDispatcher()
	err = eventDispatcher.Register("OrderCreated", &handler.OrderCreatedHandler{
		RabbitMQChannel: rabbitMQChannel,
	})
	if err != nil {
		log.Fatalf("Failed to register event handler: %v", err)
	}

	createOrderUseCase := NewCreateOrderUseCase(db, eventDispatcher)
	listOrdersUseCase := NewListOrdersUseCase(db)

	webServer := webserver.NewWebServer(config.WebServerPort)
	webOrderHandler := NewWebOrderHandler(db, eventDispatcher)
	webListOrdersHandler := NewWebListOrdersHandler(db)

	webServer.AddHandler("/order", webOrderHandler.Create, "POST")
	webServer.AddHandler("/order", webListOrdersHandler.List, "GET")
	go startWebServer(webServer, config.WebServerPort)

	go startGRPCServer(config.GRPCServerPort, createOrderUseCase, listOrdersUseCase)

	go startGraphQLServer(config.GraphQLServerPort, createOrderUseCase, listOrdersUseCase)
	waitForShutdown()
}

func initDB(config *configs.Config) *sql.DB {
	db, err := sql.Open(config.DBDriver, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

func getRabbitMQChannel(config *configs.Config) *amqp.Channel {
	connString := fmt.Sprintf("amqp://%s:%s@%s:%s/", config.RabbitMQUser, config.RabbitMQPassword, config.RabbitMQHost, config.RabbitMQPort)
	log.Println("Connecting to RabbitMQ at", connString)

	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}

	_, err = ch.QueueDeclare(
		"order_created_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	return ch
}

func startWebServer(server *webserver.WebServer, port string) {
	log.Printf("Starting web server on port %s", port)
	server.Start()
}

func startGRPCServer(port string, createOrderUseCase *usecase.CreateOrderUseCase, listOrdersUseCase *usecase.ListOrdersUseCase) {
	grpcServer := grpc.NewServer()
	orderService := service.NewOrderService(*createOrderUseCase, *listOrdersUseCase)
	pb.RegisterOrderServiceServer(grpcServer, orderService)
	reflection.Register(grpcServer)

	log.Printf("Starting gRPC server on port %s", port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}

func startGraphQLServer(port string, createOrderUseCase *usecase.CreateOrderUseCase, listOrdersUseCase *usecase.ListOrdersUseCase) {
	srv := graphqlhandler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		CreateOrderUseCase: *createOrderUseCase,
		ListOrdersUseCase:  *listOrdersUseCase,
	}}))
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("Starting GraphQL server on port %s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("GraphQL server failed: %v", err)
	}
}

func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")
	time.Sleep(2 * time.Second)
}
