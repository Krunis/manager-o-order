package apigateway

import (
	"context"
	"log"
	"time"
)

func (g *GatewayServer) startPolling() {
	timer := time.NewTimer(time.Second * 1)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			ctx, cancel := context.WithTimeout(g.lifecycle.Ctx, time.Millisecond*400)
			defer cancel()

			rows, err := g.poolDB.Query(ctx, `SELECT topic, key, payload FROM outbox
										 WHERE sent_at=NULL`)
			if err != nil {
				log.Printf("Failed to poll from outbox: %s", err)
			}

			var topic, key string

			var value []byte

			for rows.Next() {
				err := rows.Scan(&topic, &key, &value) // row from outbox
				if err != nil {
					log.Printf("Errow while scan row: %s", err)
				}

				if err := g.sendInKafka(topic, []byte(key), value); err != nil {
					log.Printf("Error while send in Kafka: %s", err)
				}
			}

		case <-g.lifecycle.Ctx.Done():
			return 
		}
	}
}
