package main

import (
	"log"

	"github.com/Krunis/manager-o-order/packages/common"
	sagacoordinator "github.com/Krunis/manager-o-order/packages/saga_coordinator"
)

func main() {
	coord := sagacoordinator.NewCoordinator(common.GetDBConnectionString())

	if err := coord.StartCoordinator(":8084", ":8083", ":8082", []string{"sagas"}); err != nil{
		log.Println(err)
	}


}