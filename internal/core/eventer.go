package core

import (
	"errors"
	"strings"

	"github.com/Koyo-os/get-getway/internal/entity"
	"github.com/bytedance/sonic"
)

type Entity any

type Client[T Entity] interface{
	Create(T) error
	Update(string, string, string) error
	Delete(string) error
}

func route[T any](client Client[T], eventType string, payload string) error {
	lastPart := strings.Split(eventType, ".")[1]

	switch lastPart{
	case "created":
		var req entity.Create[T]

		if err := sonic.Unmarshal([]byte(payload), &req);err != nil{
			return err
		}

		return client.Create(req.Payload)
	case "updated":
		var req entity.Update

		if err := sonic.Unmarshal([]byte(payload), &req);err != nil{
			return err
		}

		return client.Update(req.ID, req.Key, req.Value)
	case "deleted":
		var req entity.Delete

		if err := sonic.Unmarshal([]byte(payload), &req);err != nil{
			return err
		}

		return client.Delete(req.ID)
	default:
		return errors.New("unknown event type")
	}
}