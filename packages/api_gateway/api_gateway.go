package apigateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/Krunis/manager-o-order/packages/common"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
)

type GatewayServer struct {
	address string

	httpServer *http.Server
	mux        *http.ServeMux

	kafkaAddress   string
	saramaProducer sarama.SyncProducer

	poolDB *pgxpool.Pool

	RedisDB *redis.Client

	stopOnce sync.Once

	wg sync.WaitGroup

	lifecycle common.Lifecycle
}

func NewGatewayServer(port, kafkaAddress string) *GatewayServer {
	mux := http.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())

	return &GatewayServer{
		address:      port,
		mux:          mux,
		kafkaAddress: kafkaAddress,
		lifecycle:    common.Lifecycle{Ctx: ctx, Cancel: cancel},
	}
}

func (g *GatewayServer) Start(dbConnectionString string) error {
	var err error

	g.saramaProducer, err = NewSaramaProducer(g.kafkaAddress)
	if err != nil {
		return err
	}

	err = func() error {
		ctx, cancel := context.WithCancel(g.lifecycle.Ctx)
		defer cancel()

		g.poolDB, err = common.ConnectToDB(ctx, dbConnectionString)
		if err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		return err
	}

	g.wg.Go(g.startPolling)

	err = func() error {
		ctx, cancel := context.WithTimeout(g.lifecycle.Ctx, time.Second*5)
		defer cancel()

		g.RedisDB, err = common.ConnectToRedis(ctx)
		if err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		return err
	}

	g.mux.HandleFunc("/order/new", g.NewOrderHandler)
	g.mux.HandleFunc("/item/add", g.AddItemHandler)
	g.mux.HandleFunc("/employee/add", g.AddEmployeeHandler)

	g.httpServer = &http.Server{}

	g.httpServer.Handler = g.mux
	g.httpServer.Addr = g.address

	errCh := make(chan error, 1)

	go func() {
		log.Println("Starting listening...")
		if err := g.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		} else {
			errCh <- nil
		}
	}()

	select {
	case err := <-errCh:
		log.Printf("Error while working: %s", err)

		return g.Stop()
	case <-g.lifecycle.Ctx.Done():
		return nil
	}
}

func (g *GatewayServer) NewOrderHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case <-g.lifecycle.Ctx.Done():
		log.Println("Request cancelled (shutdown or client disconnected)")
		return
	default:
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var order *common.Order

		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("Received order: %v\n", order)

		if err := ValidateOrder(order); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("Validate success")

		ctxRedis, cancel := context.WithTimeout(r.Context(), time.Second*1)
		defer cancel()

		exist, err := g.checkInRedis(ctxRedis, order.IdempotencyKey)
		if err != nil {
			http.Error(w, "failed to check order", http.StatusInternalServerError)
			return
		}
		if exist {
			http.Error(w, "order already exists", http.StatusConflict)
			return
		}
		log.Println("Checked in Redis")

		ctxPG, cancel := context.WithTimeout(r.Context(), time.Second*1)
		defer cancel()

		id, err := g.sendOrderInPostgres(ctxPG, order)
		if err != nil {
			http.Error(w, "failed to create order", http.StatusInternalServerError)
			return
		}
		log.Println("Sent in Postgres")

		resp := struct {
			OrderId       string `json:"order_id"`
			Status        string `json:"status"`
			EstimatedTime string `json:"estimated_time"`
			StatusUrl     string `json:"status_url"`
		}{
			OrderId:       id,
			Status:        "PENDING",
			EstimatedTime: "40m",
			StatusUrl:     fmt.Sprintf("order/status?order_id=%s", id),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

		w.WriteHeader(http.StatusCreated)
	}
}

func (g *GatewayServer) AddItemHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case <-g.lifecycle.Ctx.Done():
		log.Println("Request cancelled (shutdown or client disconnected)")
		return
	default:
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		item := &common.Item{}

		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil{
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := ValidateItem(item); err != nil{
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second * 2)
		defer cancel()

		err = g.sendItemInPostgres(ctx, item)
		if err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func (g *GatewayServer) AddEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case <-g.lifecycle.Ctx.Done():
		log.Println("Request cancelled (shutdown or client disconnected)")
		return
	default:
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}
		
		employee := &common.Employee{}

		err := json.NewDecoder(r.Body).Decode(&employee)
		if err != nil{
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := ValidateEmployee(employee); err != nil{
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	
		ctx, cancel := context.WithTimeout(r.Context(), time.Second * 2)
		defer cancel()

		if err := g.sendEmployeeInPostgres(ctx, employee); err != nil{
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func (g *GatewayServer) Stop() error {
	var result error

	g.stopOnce.Do(func() {
		var errs []error

		g.lifecycle.Cancel()

		if g.httpServer != nil {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			log.Println("Graceful shutdown...")
			if err := g.httpServer.Shutdown(shutdownCtx); err != nil {
				log.Printf("Graceful shutdown failed: %s\n", err)

				if err = g.httpServer.Close(); err != nil {
					log.Printf("Force close failed: %s\n", err)
					err = fmt.Errorf("shutdown failed: %v, close failed: %v", err, err)
					return
				}
				err = fmt.Errorf("shutdown failed: %v, forced close", err)

				errs = append(errs, err)
			}
		}

		log.Println("Shutdown completed")

		if g.RedisDB != nil {
			if err := g.RedisDB.Close(); err != nil {
				errs = append(errs, err)
			}
			log.Println("Stopped connect to redis")
		}

		if g.poolDB != nil {
			g.poolDB.Close()
			log.Println("Database pool stopped")
		}

		if g.saramaProducer != nil {
			if err := g.saramaProducer.Close(); err != nil {
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			result = errors.Join(errs...)
		}

	})

	return result
}
