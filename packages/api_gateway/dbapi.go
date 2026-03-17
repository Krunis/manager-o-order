package apigateway

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func (g *GatewayServer) sendInPostgres(ctx context.Context, order *Order) error{
	tx, err := g.poolDB.Begin(ctx)
	if err != nil{
		return err
	}
	defer tx.Rollback(ctx)

	tx.Exec(ctx, `INSERT`)

	// outbox
}

func (g *GatewayServer) checkInRedis(ctx context.Context, IdempotencyKey string) (bool, error){
	_, err := g.RedisDB.Get(ctx, IdempotencyKey).Result()
	if err != nil{
		if err == redis.Nil{
			return false, nil
		}

		return true, err
	}

	return true, nil
}