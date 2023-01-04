package lbtmw

import (
	"strings"
	"io/ioutil"
	"os"
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

const RedisKeyDelimiter = ":"

var RedisScriptPath = "./redis_script/"

var redisScriptCache = make(map[string]*redis.Script)

var RedisClient redis.UniversalClient

var Host string

// 1. If the MasterName option is specified, a sentinel-backed FailoverClient is returned.
// 2. if the number of Addrs is two or more, a ClusterClient is returned.
// 3. Otherwise, a single-node Client is returned.
func InitRedisClient(ops *redis.UniversalOptions) {
	RedisClient = redis.NewUniversalClient(ops)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		panic(err)
	}
	logger.Info("init redis %v %v", ops.Addrs, ops.MasterName)
}

func RedisKey(fields []string) string {
	return strings.Join(fields, RedisKeyDelimiter)
}

func RedisHostKey(fields []string) string {
	s := strings.Join(fields, RedisKeyDelimiter)
	return strings.Join([]string{Host, s}, RedisKeyDelimiter)
}

func LoadAllRedisScript(spath string) {
	// 加载路径下所有.lua脚本
	dir, err := ioutil.ReadDir(spath)
	if err != nil {
		panic(err)
	}
	for _, file := range dir {
		if !file.IsDir() {
			s := strings.Split(file.Name(), ".")
			LoadRedisScript(spath, s[0])
		}
	}
}

func LoadRedisScript(spath string, sname string) {
	// 加载路径下单个.lua脚本
	file, err := os.Open(spath + sname + ".lua")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	content, _ := ioutil.ReadAll(file)
	c := redis.NewScript(string(content))
	redisScriptCache[sname] = c
}

func RunRedisScript(sname string, key string, args ...interface{}) (interface{}, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := redisScriptCache[sname].Eval(ctx, RedisClient, []string{key}, args...).Result()
	if err != nil {
		logger.Error("RunRedisScript error %v %v %v %v", sname, key, args, err)
		return nil, false
	}
	return r, err == nil
}
