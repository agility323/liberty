package lbtmw

import (
	"strings"

	"github.com/go-redis/redis"
)

const RedisKeyDelimiter = ":"

var redisClient redis.UniversalClient

// 1. If the MasterName option is specified, a sentinel-backed FailoverClient is returned.
// 2. if the number of Addrs is two or more, a ClusterClient is returned.
// 3. Otherwise, a single-node Client is returned.
func InitRedisClient(addrs []string, masterName string) {
	redisClient = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: addrs,
		MasterName: masterName,
	})
	if err := redisClient.Ping().Err(); err != nil {
		panic(err)
	}
	logger.Info("init redis %v %s", addrs, masterName)
}

func RedisKey(fields []string) string {
	return strings.Join(fields, RedisKeyDelimiter)
}
