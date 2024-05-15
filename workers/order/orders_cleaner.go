package order

import (
	"context"
	"gitlab.com/iaroslavtsevaleksandr/geotask/module/order/service"
	"log"
	"time"
)

const (
	orderCleanInterval = 5 * time.Second
)

// OrderCleaner воркер, который удаляет старые заказы
// используя метод orderService.RemoveOldOrders()
type OrderCleaner struct {
	orderService service.Orderer
}

func NewOrderCleaner(orderService service.Orderer) *OrderCleaner {
	return &OrderCleaner{orderService: orderService}
}

func (o *OrderCleaner) Run() {

	ticker := time.NewTicker(orderCleanInterval)
	for {
		select {
		case <-ticker.C:
			if err := o.orderService.RemoveOldOrders(context.Background()); err != nil {
				log.Printf("Failed to remove old orders: %v", err)
			}
		}
	}

	// исользовать горутину и select
	// внутри горутины нужно использовать time.NewTicker()
	// и вызывать метод orderService.RemoveOldOrders()
	// если при удалении заказов произошла ошибка, то нужно вывести ее в лог
}
