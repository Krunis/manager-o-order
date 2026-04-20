package main

import (
	"log"

	"github.com/Krunis/manager-o-order/packages/common"
	sagacoordinator "github.com/Krunis/manager-o-order/packages/saga_coordinator"
)

func main() {
	coord := sagacoordinator.NewCoordinator(common.GetDBConnectionString())

	if err := coord.StartCoordinator("confirmation:8084", "storage:8083", "delivery:8082", []string{"orders"}); err != nil{
		log.Println(err)
	}


}