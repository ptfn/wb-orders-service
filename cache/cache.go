package cache

import (
	"sync"
	"wb-orders-service/models"
)

type Cache struct {
	mu    sync.RWMutex
	data  map[string]*models.Order
}

func New() *Cache {
	return &Cache{
		data: make(map[string]*models.Order),
	}
}

// Get возвращает заказ по order_uid
func (c *Cache) Get(orderUID string) (*models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, exists := c.data[orderUID]
	return order, exists
}

// Set сохраняет заказ в кэш
func (c *Cache) Set(order *models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[order.OrderUID] = order
}

// GetAll возвращает все заказы в кэше (для отладки)
func (c *Cache) GetAll() map[string]*models.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Создаем копию чтобы избежать гонок данных
	result := make(map[string]*models.Order)
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// Size возвращает количество элементов в кэше
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.data)
}

// Delete удаляет заказ из кэша
func (c *Cache) Delete(orderUID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, orderUID)
}
