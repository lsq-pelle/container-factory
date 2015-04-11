package thingamabob

import (
	"errors"
	"math"

	"gopkg.in/redis.v2"
)

type Client struct {
	redis *redis.Client
	keys  [2]string
}

type Options struct {
	Redis  *redis.Client
	Bucket string
}

var (
	NotFound = errors.New("thingamabob: ID not found")
	Invalid  = errors.New("thingamabob: invalid ID")
)

var (
	noBucket = errors.New("thingamabob: a bucket name must be supplied")
	noRedis  = errors.New("thingamabob: a Redis client must be supplied")
)

func NewClient(opts *Options) (*Client, error) {
	if opts.Bucket == "" {
		return nil, noBucket
	}

	if opts.Redis == nil {
		return nil, noRedis
	}

	if err := opts.Redis.Ping().Err(); err != nil {
		return nil, err
	}

	return &Client{
		redis: opts.Redis,
		keys: [2]string{
			opts.Bucket + "/list",
			opts.Bucket + "/index",
		},
	}, nil
}

var idScript = newScript(`
	local list = KEYS[1]
	local index = KEYS[2]

	local name = ARGV[1]

	local id = redis.call('HGET', index, name)
	if id then
		return tonumber(id)
	end

	id = redis.call('LPUSH', list, name) - 1
	redis.call('HSET', index, name, id)

	return id`)

func (c *Client) Id(name string) (id uint64, err error) {
	idResult, err := idScript.eval(c.redis, c.keys[:], []string{name}).Result()

	if err != nil {
		return math.MaxUint64, err
	}

	return uint64(idResult.(int64)), nil
}

func (c *Client) Name(id uint64) (name string, err error) {
	if id > uint64(math.MaxInt64) {
		return "", Invalid
	}

	name, err = c.redis.LIndex(c.keys[0], int64(id)).Result()
	return
}

func (c *Client) Destroy() error {
	return c.redis.Del(c.keys[:]...).Err()
}
