package main

import (
	"log"
	"os"

	"github.com/LiveSqrd/thingamabob"
	"gopkg.in/redis.v2"
)

var nameRegistry *thingamabob.Client

func init() {
	var err error
	nameRegistry, err = thingamabob.NewClient(&thingamabob.Options{
		Redis: redis.NewClient(&redis.Options{
			Network:  "tcp",
			Addr:     "redis-thingamabob.lsqio.svc.tutum.io:8743",
			Password: os.Getenv("THINGAMABOB_TOKEN"),
		}),
		Bucket: "docker-image-names",
	})

	if err != nil {
		log.Fatal("container-factory: can't connect to Thingamabob: ", err)
	}
}
