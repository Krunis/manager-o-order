package main

import (
	"log"

	"github.com/Krunis/manager-o-order/packages/common"
	"github.com/Krunis/manager-o-order/packages/delivery"
)

func main() {
	delv := delivery.NewDeliveryService(common.GetDBConnectionString())

	if err := delv.Start(":8082"); err != nil {
		log.Println(err)
	}
}