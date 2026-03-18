package apigateway

import (
	"context"
	"time"
)

func (g *GatewayServer) startPolling() error {
	timer := time.NewTimer(time.Second * 1)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			ctx, cancel := context.WithTimeout(g.lifecycle.Ctx, time.Millisecond * 400)
			defer cancel()

			rows, err := g.poolDB.Query(ctx, `SELECT * FROM outbox
										 WHERE sent_at=NULL`)
			if err != nil{
				return err
			}

			for rows.Next(){
				err := rows.Scan() // row from outbox
			}

			if err := g.sendInKafka(); err != nil {
				http.Error(w, "unknown error", http.StatusInternalServerError)
				return
			}
		case <-g.lifecycle.Ctx.Done():
		}
	}
}
