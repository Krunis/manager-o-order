package main

import (
	"log"

	"github.com/Krunis/manager-o-order/packages/common"
	"github.com/Krunis/manager-o-order/packages/storage"
)

func main() {
	stor := storage.NewStorageService(common.GetDBConnectionString())

	if err := stor.Start(":8083"); err != nil {
		log.Println(err)
	}
}