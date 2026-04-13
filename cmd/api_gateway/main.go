package main

import (
	"log"

	apigateway "github.com/Krunis/manager-o-order/packages/api_gateway"
	"github.com/Krunis/manager-o-order/packages/common"
)

func main() {
	gate := apigateway.NewGatewayServer(":8080", "kafka:9092")

	if err := gate.Start(common.GetDBConnectionString()); err != nil{
		log.Println(err)
	}


}