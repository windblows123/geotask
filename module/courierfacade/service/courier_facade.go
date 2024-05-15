package service

import (
	"context"
	cservice "gitlab.com/iaroslavtsevaleksandr/geotask/module/courier/service"
	cfm "gitlab.com/iaroslavtsevaleksandr/geotask/module/courierfacade/models"
	oservice "gitlab.com/iaroslavtsevaleksandr/geotask/module/order/service"
	"log"
)

const (
	CourierVisibilityRadius = 2800 // 2.8km
)

type CourierFacer interface {
	MoveCourier(ctx context.Context, direction, zoom int) // отвечает за движение курьера по карте direction - направление движения, zoom - уровень зума
	GetStatus(ctx context.Context) cfm.CourierStatus      // отвечает за получение статуса курьера и заказов вокруг него
}

// CourierFacade фасад для курьера и заказов вокруг него (для фронта)
type CourierFacade struct {
	courierService cservice.Courierer
	orderService   oservice.Orderer
}

func (c CourierFacade) MoveCourier(ctx context.Context, direction, zoom int) {
	//TODO implement me
	panic("implement me")
}

func (c CourierFacade) GetStatus(ctx context.Context) cfm.CourierStatus {
	status := cfm.CourierStatus{}
	courier, err := c.courierService.GetCourier(ctx)
	if err != nil {
		log.Println("Error getting courier in GetStatus:", err)
		return status
	}
	orders, err := c.orderService.GetByRadius(ctx, courier.Location.Lng, courier.Location.Lat, CourierVisibilityRadius, "km")
	if err != nil {
		log.Println("Error getting orders in GetStatus:", err)
		return status
	}
	status.Courier = *courier
	status.Orders = orders
	return status
}

func NewCourierFacade(courierService cservice.Courierer, orderService oservice.Orderer) CourierFacer {
	return &CourierFacade{courierService: courierService, orderService: orderService}
}
