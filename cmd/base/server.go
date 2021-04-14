package main

import (
	"flag"
	"github.com/yametech/verthandi/pkg/api"
	"github.com/yametech/verthandi/pkg/api/action/base"
	"github.com/yametech/verthandi/pkg/service"
	"github.com/yametech/verthandi/pkg/store/mongo"
)

var storageUri string

func main() {
	flag.StringVar(&storageUri, "storage_uri", "mongodb://0.0.0.0:27017/admin", "127.0.0.1:3306")
	flag.Parse()

	//errC := make(chan error)
	//store, err := mysql.Setup(storageUri, user, pw, database, errC)
	store, err, errC := mongo.NewMongo(storageUri)
	if err != nil {
		panic(err)
	}

	baseService := service.NewBaseService(store)
	server := api.NewServer(baseService)
	base.NewBaseServer("baseserver", server)

	go func() {
		if err := server.Run(":8080"); err != nil {
			errC <- err
		}
	}()

	if e := <-errC; e != nil {
		panic(e)
	}

}
