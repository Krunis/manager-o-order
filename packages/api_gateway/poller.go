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

			func() {
				tx, err := g.poolDB.Begin(ctx)
				if err != nil {
					log.Println(err)
				}
				defer tx.Rollback(ctx)

				rows, err := tx.Query(ctx, `UPDATE outbox
										SET sent_at=NOW()
										WHERE idemp_key IN (
											SELECT idemp_key
											FROM outbox
											WHERE sent_at IS NULL
											LIMIT 100
											)
										RETURNING topic, key, payload`)
				if err != nil {
					log.Printf("Failed to poll from outbox: %s", err)
				}

				var topic, key string

				var value []byte

				for rows.Next() {
					err := rows.Scan(&topic, &key, &value)
					if err != nil {
						log.Printf("Errow while scan row: %s", err)
					}

					if err := g.sendInKafka(topic, []byte(key), value); err != nil {
						log.Printf("Error while send in Kafka: %s", err)
					}
				}

				tx.Commit(ctx)
			}()

		case <-g.lifecycle.Ctx.Done():
			return
		}
	}
}
