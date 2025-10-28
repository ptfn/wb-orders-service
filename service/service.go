package service

import (
	"log"
	"wb-orders-service/cache"
	"wb-orders-service/models"
	"wb-orders-service/repository"
)

type OrderService struct {
	repo  *repository.PostgresRepository
	cache *cache.Cache
}

func NewOrderService(repo *repository.PostgresRepository) *OrderService {
	service := &OrderService{
		repo:  repo,
		cache: cache.New(),
	}

	// Восстанавливаем кэш из БД при создании сервиса
	if err := service.restoreCache(); err != nil {
		log.Printf("Warning: failed to restore cache: %v", err)
	}

	return service
}

// restoreCache загружает все заказы из БД в кэш
func (s *OrderService) restoreCache() error {
	orders, err := s.repo.GetAllOrders()
	if err != nil {
		return err
	}

	for i := range orders {
		s.cache.Set(&orders[i])
	}

	log.Printf("Cache restored with %d orders", len(orders))
	return nil
}

// SaveOrder сохраняет заказ в БД и обновляет кэш
func (s *OrderService) SaveOrder(order *models.Order) error {
	// Сохраняем в БД
	if err := s.repo.SaveOrder(order); err != nil {
		return err
	}

	// Обновляем кэш
	s.cache.Set(order)

	log.Printf("Order %s saved to DB and cache", order.OrderUID)
	return nil
}

// GetOrder возвращает заказ из кэша или БД
func (s *OrderService) GetOrder(orderUID string) (*models.Order, error) {
	// Пробуем получить из кэша (быстро)
	if order, exists := s.cache.Get(orderUID); exists {
		log.Printf("Order %s found in cache", orderUID)
		return order, nil
	}

	// Если нет в кэше, ищем в БД
	order, err := s.repo.GetOrderByUID(orderUID)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш для будущих запросов
	s.cache.Set(order)
	log.Printf("Order %s loaded from DB and cached", orderUID)

	return order, nil
}

// GetCacheSize возвращает размер кэша
func (s *OrderService) GetCacheSize() int {
	return s.cache.Size()
}
