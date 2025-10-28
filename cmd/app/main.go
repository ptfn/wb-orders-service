package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"wb-orders-service/config"
	"wb-orders-service/httpserver"
	"wb-orders-service/nats"
	"wb-orders-service/repository"
	"wb-orders-service/service"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Подключаемся к БД
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	repo, err := repository.NewPostgresRepository(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repo.Close()

	log.Println("Successfully connected to database")

	// Создаем сервис (автоматически восстанавливает кэш из БД)
	orderService := service.NewOrderService(repo)
	
	log.Printf("Service started with %d orders in cache", orderService.GetCacheSize())

	// Создаем и настраиваем NATS подписчика
	subscriber := nats.NewSubscriber(
		cfg.NATS.ClusterID,
		cfg.NATS.ClientID,
		cfg.NATS.URL,
		cfg.NATS.Subject,
		orderService,
	)

	// Подключаемся к NATS
	if err := subscriber.Connect(); err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer subscriber.Close()

	// Подписываемся на канал
	subscription, err := subscriber.Subscribe()
	if err != nil {
		log.Fatalf("Failed to subscribe to NATS: %v", err)
	}
	defer subscription.Unsubscribe()

	log.Println("NATS subscriber started successfully")

	// Создаем HTTP роутер
	router := httpserver.NewRouter(orderService)

	// Запускаем HTTP сервер
	server := &http.Server{
		Addr:    ":" + cfg.HTTP.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Starting HTTP server on port %s", cfg.HTTP.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Ожидаем сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	log.Println("Service is running. Press Ctrl+C to stop.")
	log.Printf("Web interface: http://localhost:%s", cfg.HTTP.Port)
	log.Printf("Health check: http://localhost:%s/health", cfg.HTTP.Port)
	log.Printf("Get order: http://localhost:%s/order/{id}", cfg.HTTP.Port)

	<-sigChan
	log.Println("Shutting down service...")
}
