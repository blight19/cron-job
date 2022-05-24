package service

import (
	"golang.org/x/net/context"
)

type HelloService struct {
}

func (s *HelloService) GetOrder(ctx context.Context, in *OrderId) (*Order, error) {
	// 服务实现.

	order := Order{
		Id:    in.Id,
		Items: []string{"aaa", "bb"},
		Price: 100,
	}
	return &order, nil
}
