package thingamabob

import "gopkg.in/redis.v2"

func getClient() (client *Client) {
	var err error
	redisClient := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "localhost:6379",
	})
	client, err = NewClient(&Options{
		Bucket: "test",
		Redis:  redisClient,
	})
	if err != nil {
		panic(err)
	}

	if err = client.Destroy(); err != nil {
		panic(err)
	}

	return
}
