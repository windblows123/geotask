package service

import (
	"context"
	"gitlab.com/iaroslavtsevaleksandr/geotask/geo"
	"gitlab.com/iaroslavtsevaleksandr/geotask/module/courier/models"
	"gitlab.com/iaroslavtsevaleksandr/geotask/module/courier/storage"
	"log"
)

// Направления движения курьера
const (
	DirectionUp    = 0
	DirectionDown  = 1
	DirectionLeft  = 2
	DirectionRight = 3
)

const (
	DefaultCourierLat = 59.9311
	DefaultCourierLng = 30.3609
)

type Courierer interface {
	GetCourier(ctx context.Context) (*models.Courier, error)
	MoveCourier(courier models.Courier, direction, zoom int) error
}

type CourierService struct {
	courierStorage storage.CourierStorager
	allowedZone    geo.PolygonChecker
	disabledZones  []geo.PolygonChecker
}

func NewCourierService(courierStorage storage.CourierStorager, allowedZone geo.PolygonChecker, disbledZones []geo.PolygonChecker) Courierer {
	return &CourierService{courierStorage: courierStorage, allowedZone: allowedZone, disabledZones: disbledZones}
}

func (c *CourierService) GetCourier(ctx context.Context) (*models.Courier, error) {
	// получаем курьера из хранилища используя метод GetOne из storage/courier.go
	courier, err := c.courierStorage.GetOne(ctx)
	if err != nil {
		log.Println("error getting courier in service GetCourier:", err)
		return nil, err
	}

	// проверяем, что курьер находится в разрешенной зоне
	// если нет, то перемещаем его в случайную точку в разрешенной зоне
	if c.allowedZone.Contains(geo.Point{
		Lat: courier.Location.Lat,
		Lng: courier.Location.Lng,
	}) {
		if err := c.courierStorage.Save(ctx, *courier); err != nil {
			log.Println("error saving courier in service GetCourier:", err)
		}
		return courier, nil
	} else {
		point := c.allowedZone.RandomPoint()
		courier.Location.Lng = point.Lng
		courier.Location.Lat = point.Lat
		if err := c.courierStorage.Save(ctx, *courier); err != nil {
			log.Println("error saving courier in service GetCourier:", err)
		}
		return courier, nil
	}
	// сохраняем новые координаты курьера
}

// MoveCourier : direction - направление движения курьера, zoom - зум карты
func (c *CourierService) MoveCourier(courier models.Courier, direction, zoom int) error {
	// точность перемещения зависит от зума карты использовать формулу 0.001 / 2^(zoom - 14)
	// 14 - это максимальный зум карты

	// далее нужно проверить, что курьер не вышел за границы зоны
	// если вышел, то нужно переместить его в случайную точку внутри зоны

	// далее сохранить изменения в хранилище
	return nil
}
