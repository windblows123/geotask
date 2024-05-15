package storage

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis"
	"gitlab.com/iaroslavtsevaleksandr/geotask/module/courier/models"
	"log"
)

type CourierStorager interface {
	Save(ctx context.Context, courier models.Courier) error // сохранить курьера по ключу courier
	GetOne(ctx context.Context) (*models.Courier, error)    // получить курьера по ключу courier
}

type CourierStorage struct {
	storage *redis.Client
}

func (c *CourierStorage) Save(ctx context.Context, courier models.Courier) error {
	data, err := json.Marshal(courier)
	if err != nil {
		log.Println("error marshalling courier location")
		return err
	}
	c.storage.ZAdd("courier", redis.Z{
		Score:  float64(courier.Score),
		Member: string(data),
	})
	return nil
}

func (c *CourierStorage) GetOne(ctx context.Context) (*models.Courier, error) {
	var courier models.Courier
	//rand.Seed(time.Now().Unix())
	//members, err := c.storage.ZRangeByScore("courier", redis.ZRangeBy{
	//	Min:    "0",
	//	Max:    "6",
	//}).Result()
	//if err != nil {
	//	return nil, err
	//}
	//index := rand.Intn(len(members))
	//if err := json.Unmarshal([]byte(members[index]), &courier); err != nil {
	//	log.Println("error unmarshalling courier")
	//	return nil, err
	//}
	members, err := c.storage.ZPopMax("1").Result()
	if err != nil {
		log.Println("error popping courier from redis")
		return nil, err
	}
	courierJSON, ok := members[0].Member.(string)
	if !ok {
		log.Println("error marshalling courier location")
		return nil, err
	}
	if err := json.Unmarshal([]byte(courierJSON), &courier); err != nil {
		log.Println("error unmarshalling courier location")
		return nil, err
	}
	return &courier, nil
	//members, err := c.storage.ZRandMember(ctx, "courier", 1).Result()
	//if err != nil {
	//	return nil, err
	//}
	//if err := json.Unmarshal([]byte(members[0]), &courier); err != nil {
	//	log.Println("error unmarshalling courier")
	//	return nil, err
	//}
}

func NewCourierStorage(storage *redis.Client) CourierStorager {
	return &CourierStorage{storage: storage}
}
