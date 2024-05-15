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
		randomPoint := c.allowedZone.RandomPoint()
		courier.Location.Lng = randomPoint.Lng
		courier.Location.Lat = randomPoint.Lat
		if err := c.courierStorage.Save(ctx, *courier); err != nil {
			log.Println("error saving courier in service GetCourier:", err)
		}
		return courier, nil
	}
	// сохраняем новые координаты курьера
}

// MoveCourier : direction - направление движения курьера, zoom - зум карты
func (c *CourierService) MoveCourier(courier models.Courier, direction, zoom int) error {
	// Рассчитываем точность перемещения в зависимости от уровня зума карты
	stepSize := 0.001 / (1 << uint(14-zoom))

	// Вычисляем новые координаты курьера в зависимости от направления и шага перемещения
	switch direction {
	case DirectionUp:
		courier.Location.Lat += stepSize
	case DirectionDown:
		courier.Location.Lat -= stepSize
	case DirectionLeft:
		courier.Location.Lng -= stepSize
	case DirectionRight:
		courier.Location.Lng += stepSize
	}

	// Проверяем, что курьер остается в пределах разрешенной зоны
	point := geo.Point{
		Lat: courier.Location.Lat,
		Lng: courier.Location.Lng,
	}
	if !c.allowedZone.Contains(point) {
		// Курьер вышел за пределы разрешенной зоны, перемещаем его в случайную точку внутри разрешенной зоны
		randomPoint := c.allowedZone.RandomPoint()
		courier.Location.Lng = randomPoint.Lng
		courier.Location.Lat = randomPoint.Lat
	}

	// Сохраняем изменения в хранилище
	err := c.courierStorage.Save(context.Background(), courier)
	if err != nil {
		return err
	}

	return nil
}

func (c *CourierService) getRandomAllowedPoint() (float64, float64) {
	return DefaultCourierLat, DefaultCourierLng
}
