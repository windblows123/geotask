package models

import (
	cm "gitlab.com/iaroslavtsevaleksandr/geotask/module/courier/models"
	om "gitlab.com/iaroslavtsevaleksandr/geotask/module/order/models"
)

type CourierStatus struct {
	Courier cm.Courier `json:"courier"`
	Orders  []om.Order `json:"orders"`
}
