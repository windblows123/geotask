package storage

//TODO разобратьсся с версией редиса
import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.com/iaroslavtsevaleksandr/geotask/module/order/models"
	"log"
	"strconv"
	"time"
)

type OrderStorager interface {
	Save(ctx context.Context, order models.Order, maxAge time.Duration) error                       // сохранить заказ с временем жизни
	GetByID(ctx context.Context, orderID int) (*models.Order, error)                                // получить заказ по id
	GenerateUniqueID(ctx context.Context) (int64, error)                                            // сгенерировать уникальный id
	GetByRadius(ctx context.Context, lng, lat, radius float64, unit string) ([]models.Order, error) // получить заказы в радиусе от точки
	GetCount(ctx context.Context) (int, error)                                                      // получить количество заказов
	RemoveOldOrders(ctx context.Context, maxAge time.Duration) error                                // удалить старые заказы по истечению времени maxAge
}

type OrderStorage struct {
	storage *redis.Client
}

func NewOrderStorage(storage *redis.Client) OrderStorager {
	return &OrderStorage{storage: storage}
}

func (o *OrderStorage) Save(ctx context.Context, order models.Order, maxAge time.Duration) error {
	// save with geo redis
	return o.saveOrderWithGeo(ctx, order, maxAge)
}

func (o *OrderStorage) RemoveOldOrders(ctx context.Context, maxAge time.Duration) error {
	// получить ID всех старых ордеров, которые нужно удалить
	// используя метод ZRangeByScore
	// старые ордеры это те, которые были созданы две минуты назад
	// и более
	/**
	&redis.ZRangeBy{
		Max: использовать вычисление времени с помощью maxAge,
		Min: "0",
	}
	*/
	if err := o.storage.ZRemRangeByScore("orders", "0", strconv.Itoa(int(time.Now().Add(-maxAge).Unix()))).
		Err(); err != nil {
		log.Printf("Failed to remove old orders: %v", err)
		return err
	}
	//ANOTHER WAY OF DOING THE SAME
	//members, err := o.storage.ZRangeByScore("orders", redis.ZRangeBy{
	//	Min: "0",
	//	Max: strconv.Itoa(int(time.Now().Add(-maxAge).Unix())),
	//}).Result()
	//if err != nil {
	//	log.Println("Failed to get members of expired orders", err)
	//	return err
	//}
	//for _, member := range members {
	//	if err := o.storage.ZRem("orders", member).Err(); err != nil {
	//		log.Printf("Failed to remove old order, member: %s, err: %v\n", members, err)
	//		return err
	//	}
	//}

	// Проверить количество старых ордеров
	// удалить старые ордеры из redis используя метод ZRemRangeByScore где ключ "orders" min "-inf" max "(время создания старого ордера)"
	// удалять ордера по ключу не нужно, они будут удалены автоматически по истечению времени жизни
	return nil
}

func (o *OrderStorage) GetByID(ctx context.Context, orderID int) (*models.Order, error) {
	var err error
	var data []byte
	var order models.Order
	orderKey := fmt.Sprintf("order:%d", orderID)
	// получаем ордер из redis по ключу order:ID
	data, err = o.storage.Get(orderKey).Bytes()
	// проверяем что ордер не найден исключение redis.Nil, в этом случае возвращаем nil, nil
	if err != nil {
		if err != redis.Nil {
			return nil, err
		}
		return nil, nil
	}
	// десериализуем ордер из json
	if err = json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (o *OrderStorage) saveOrderWithGeo(ctx context.Context, order models.Order, maxAge time.Duration) error {
	var err error
	var data []byte
	orderKey := fmt.Sprintf("order:%d", order.ID)

	// сериализуем ордер в json
	data, err = json.Marshal(order)
	// сохраняем ордер в json redis по ключу order:ID с временем жизни maxAge
	o.storage.Set(orderKey, data, maxAge)

	// добавляем ордер в гео индекс используя метод GeoAdd где Name - это ключ ордера, а Longitude и Latitude - координаты
	o.storage.GeoAdd("orders:location", &redis.GeoLocation{
		Name:      orderKey,
		Longitude: order.Lng,
		Latitude:  order.Lat,
	})

	// zset сохраняем ордер для получения количества заказов со сложностью O(1)
	// Score - время создания ордера
	o.storage.ZAdd("orders", redis.Z{
		Score:  float64(order.CreatedAt.Unix()),
		Member: data,
	})
	return err
}

func (o *OrderStorage) GetCount(ctx context.Context) (int, error) {
	// получить количество ордеров в упорядоченном множестве используя метод ZCard
	count, err := o.storage.ZCard("orders").Result()
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (o *OrderStorage) GetByRadius(ctx context.Context, lng, lat, radius float64, unit string) ([]models.Order, error) {
	var err error
	var orders []models.Order
	var data []byte
	var ordersLocation []redis.GeoLocation

	// используем метод getOrdersByRadius для получения ID заказов в радиусе
	ordersLocation, err = o.getOrdersByRadius(ctx, lng, lat, radius, unit)
	// обратите внимание, что в случае отсутствия заказов в радиусе
	// метод getOrdersByRadius должен вернуть nil, nil (при ошибке redis.Nil)
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	orders = make([]models.Order, 0, len(ordersLocation))
	// проходим по списку ID заказов и получаем данные о заказе
	// получаем данные о заказе по ID из redis по ключу order:ID
	for _, location := range ordersLocation {
		data, err = o.storage.Get(location.Name).Bytes()
		if err != nil {
			log.Println("error getting orders from storage:", err)
			return nil, err
		}
		var order models.Order
		if err = json.Unmarshal(data, &order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (o *OrderStorage) getOrdersByRadius(ctx context.Context, lng, lat, radius float64, unit string) ([]redis.GeoLocation, error) {
	// в данном методе мы получаем список ордеров в радиусе от точки
	// возвращаем список ордеров с координатами и расстоянием до точки
	/**
	&redis.GeoRadiusQuery{
		Radius:      radius,
		Unit:        unit,
		WithCoord:   true,
		WithDist:    true,
		WithGeoHash: true,
	}
	*/
	return o.storage.GeoRadius("orders:location", lng, lat, &redis.GeoRadiusQuery{
		Radius:      radius,
		Unit:        unit,
		WithCoord:   true,
		WithDist:    true,
		WithGeoHash: true,
	}).Result()
}

func (o *OrderStorage) GenerateUniqueID(ctx context.Context) (int64, error) {
	var err error
	var id int64
	id, err = o.storage.Incr("order:id").Result()
	if err != nil {
		return 0, err
	}
	// генерируем уникальный ID для ордера
	// используем для этого redis incr по ключу order:id
	return id, nil
}
