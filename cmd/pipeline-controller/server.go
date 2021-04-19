package main

import (
	"flag"
	"github.com/yametech/verthandi/pkg/controller"
	"github.com/yametech/verthandi/pkg/store/mongo"
)

var storageUri string

func main() {
	flag.StringVar(&storageUri, "storage_uri", "mongodb://127.0.0.1:27017/admin", "-storage_uri mongodb://127.0.0.1:27017/admin")
	flag.Parse()

	stage, err, errC := mongo.NewMongo(storageUri)
	if err != nil {
		panic(err)
	}

	go func() {
		if err := controller.NewPipelineController(stage).Run(); err != nil {
			errC <- err
		}
	}()

	panic(<-errC)

}
