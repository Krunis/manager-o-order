package main

import (
	"log"

	"github.com/Krunis/manager-o-order/packages/common"
	"github.com/Krunis/manager-o-order/packages/confirmation"
)

func main() {
	conf := confirmation.NewConfirmationService(common.GetDBConnectionString())

	if err := conf.Start(":8084"); err != nil{
		log.Println(err)
	}
}