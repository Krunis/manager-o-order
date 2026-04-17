package apigateway

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Krunis/manager-o-order/packages/common"
	"github.com/redis/go-redis/v9"
)

func (g *GatewayServer) sendInPostgres(ctx context.Context, order *common.Order) (string, error) {
	tx, err := g.poolDB.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var id string

	err = tx.QueryRow(ctx, `INSERT INTO orders(idemp_key, employee_id, department_id, status, confirmation_employee_id)
						   VALUES($1, $2, $3, $4, $5)
						   RETURNS id`,	
						   order.IdempotencyKey, order.EmployeeID, order.DepartmentID, "PENDING", order.ConfirmationEmployeeID).Scan(&id)
	if err != nil{
		return "", err
	}

	log.Println("Inserted in ORDERS")

	key := order.EmployeeID + order.DepartmentID

	payload, err := json.Marshal(order)
	if err != nil{
		return "", err
	}

	_, err = tx.Exec(ctx, `INSERT INTO outbox(idemp_key, topic, key, payload)
						   VALUES($1, $2, $3, $4)`,
						   order.IdempotencyKey, "orders", key, payload)
	if err != nil{
		return "", err
	}

	log.Println("Inserted in OUTBOX")
	
	return id, tx.Commit(ctx)

}

func (g *GatewayServer) checkInRedis(ctx context.Context, IdempotencyKey string) (bool, error) {
	_, err := g.RedisDB.Get(ctx, IdempotencyKey).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}

		return true, err
	}

	return true, nil
}
