package nats

import (
	"encoding/json"
	"log"
	"fmt"
	"wb-orders-service/models"
	"wb-orders-service/service"

	"github.com/nats-io/stan.go"
)

type Subscriber struct {
	clusterID string
	clientID  string
	url       string
	subject   string
	service   *service.OrderService
	conn      stan.Conn
}

func NewSubscriber(clusterID, clientID, url, subject string, service *service.OrderService) *Subscriber {
	return &Subscriber{
		clusterID: clusterID,
		clientID:  clientID,
		url:       url,
		subject:   subject,
		service:   service,
	}
}

// Connect подключается к NATS Streaming
func (s *Subscriber) Connect() error {
	conn, err := stan.Connect(s.clusterID, s.clientID, stan.NatsURL(s.url))
	if err != nil {
		return err
	}
	s.conn = conn
	log.Printf("Connected to NATS Streaming: %s", s.url)
	return nil
}

// Subscribe подписывается на канал и обрабатывает сообщения
func (s *Subscriber) Subscribe() (stan.Subscription, error) {
	subscription, err := s.conn.Subscribe(s.subject, func(msg *stan.Msg) {
		if err := s.processMessage(msg); err != nil {
			log.Printf("Failed to process message: %v", err)
		}
	}, stan.DeliverAllAvailable())

	if err != nil {
		return nil, err
	}

	log.Printf("Subscribed to subject: %s", s.subject)
	return subscription, nil
}

// processMessage обрабатывает входящее сообщение
func (s *Subscriber) processMessage(msg *stan.Msg) error {
	log.Printf("Received message: %s", string(msg.Data))

	var order models.Order
	if err := json.Unmarshal(msg.Data, &order); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return err
	}

	// Валидация данных
	if err := s.validateOrder(&order); err != nil {
		log.Printf("Order validation failed: %v", err)
		return err
	}

	// Сохраняем заказ через сервис
	if err := s.service.SaveOrder(&order); err != nil {
		log.Printf("Failed to save order: %v", err)
		return err
	}

	log.Printf("Order %s processed successfully", order.OrderUID)
	return nil
}

// validateOrder проверяет корректность данных заказа
func (s *Subscriber) validateOrder(order *models.Order) error {
	if order.OrderUID == "" {
		return fmt.Errorf("order_uid is required")
	}
	if order.TrackNumber == "" {
		return fmt.Errorf("track_number is required")
	}
	if order.Entry == "" {
		return fmt.Errorf("entry is required")
	}
	if order.Payment.Transaction == "" {
		return fmt.Errorf("payment.transaction is required")
	}
	if order.Payment.Transaction != order.OrderUID {
		return fmt.Errorf("payment.transaction must match order_uid")
	}
	return nil
}

// Close закрывает соединение
func (s *Subscriber) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}
