package apigateway

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/Krunis/manager-o-order/packages/common"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/redis/go-redis/v9"
)

func (g *GatewayServer) sendOrderInPostgres(ctx context.Context, order *common.Order) (string, error) {
	tx, err := g.poolDB.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var id string

	tag, err := tx.Exec(ctx, `
						INSERT INTO employees(id)
						VALUES($1)
						`, order.ConfirmationEmployeeID)
	if err != nil{
		return "", err
	}
	if tag.RowsAffected() == 0 {
		log.Println("Employee already in DB")
	}

	err = tx.QueryRow(ctx, `INSERT INTO orders(idemp_key, employee_id, department_id, status, confirmation_employee_id)
						   VALUES($1, $2, $3, $4, $5)
						   RETURNING id`,	
						   order.IdempotencyKey, order.EmployeeID, order.DepartmentID, "PENDING", order.ConfirmationEmployeeID).Scan(&id)
	if err != nil{
		return "", err
	}
	log.Println("Inserted in ORDERS")

	order.ID = id

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

func (g *GatewayServer) sendItemInPostgres(ctx context.Context, item *common.Item) error{
	tag, err := g.poolDB.Exec(ctx, `
	INSERT INTO storage(item_id, item_name, count, confirmation_type)
	VALUES($1, $2, $3, $4)`)
	if err != nil{
		// if err.(*pgconn.PgError).Code == pgerrcode.Dupl
		return err
	}
	if tag.RowsAffected() == 0{
		return errors.New("нихуя не поменялось почему-то")
	}
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
