package service

import (
	"context"
	"gitlab.com/iaroslavtsevaleksandr/geotask/geo"
	"gitlab.com/iaroslavtsevaleksandr/geotask/module/order/models"
	"gitlab.com/iaroslavtsevaleksandr/geotask/module/order/storage"
	"log"
	"math/rand"
	"time"
)

const (
	minDeliveryPrice = 100.00
	maxDeliveryPrice = 500.00

	maxOrderPrice = 3000.00
	minOrderPrice = 1000.00

	orderMaxAge = 2 * time.Minute
)

type Orderer interface {
	GetByRadius(ctx context.Context, lng, lat, radius float64, unit string) ([]models.Order, error) // возвращает заказы через метод storage.GetByRadius
	Save(ctx context.Context, order models.Order) error                                             // сохраняет заказ через метод storage.Save с заданным временем жизни OrderMaxAge
	GetCount(ctx context.Context) (int, error)                                                      // возвращает количество заказов через метод storage.GetCount
	RemoveOldOrders(ctx context.Context) error                                                      // удаляет старые заказы через метод storage.RemoveOldOrders с заданным временем жизни OrderMaxAge
	GenerateOrder(ctx context.Context) error                                                        // генерирует заказ в случайной точке из разрешенной зоны, с уникальным id, ценой и ценой доставки
}

// OrderService реализация интерфейса Orderer
// в нем должны быть методы GetByRadius, Save, GetCount, RemoveOldOrders, GenerateOrder
// данный сервис отвечает за работу с заказами
type OrderService struct {
	storage       storage.OrderStorager
	allowedZone   geo.PolygonChecker
	disabledZones []geo.PolygonChecker
}

func (o OrderService) GetByRadius(ctx context.Context, lng, lat, radius float64, unit string) ([]models.Order, error) {
	return o.storage.GetByRadius(ctx, lng, lat, radius, unit)
}

func (o OrderService) Save(ctx context.Context, order models.Order) error {
	return o.storage.Save(ctx, order, orderMaxAge)
}

func (o OrderService) GetCount(ctx context.Context) (int, error) {
	return o.storage.GetCount(ctx)
}

func (o OrderService) RemoveOldOrders(ctx context.Context) error {
	return o.storage.RemoveOldOrders(ctx, orderMaxAge)
}

func (o OrderService) GenerateOrder(ctx context.Context) error {
	rand.Seed(time.Now().UnixNano())
	id, err := o.storage.GenerateUniqueID(ctx)
	if err != nil {
		log.Println("error generating unique id in GenerateOrder()", err)
		return err
	}
	price := minOrderPrice + (maxOrderPrice-minOrderPrice)*rand.Float64()
	deliveryPrice := minDeliveryPrice + (maxDeliveryPrice-minDeliveryPrice)*rand.Float64()
	point := o.allowedZone.RandomPoint()
	order := models.Order{
		ID:            id,
		Price:         price,
		DeliveryPrice: deliveryPrice,
		Lng:           point.Lng,
		Lat:           point.Lat,
		IsDelivered:   false,
		CreatedAt:     time.Now(),
	}
	err = o.storage.Save(ctx, order, orderMaxAge)
	return err
}

func NewOrderService(storage storage.OrderStorager, allowedZone geo.PolygonChecker, disallowedZone []geo.PolygonChecker) Orderer {
	return &OrderService{storage: storage, allowedZone: allowedZone, disabledZones: disallowedZone}
}
