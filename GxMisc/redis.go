package GxMisc

/**
作者： Kyle Ding
模块：redis连接池
说明：
创建时间：2015-10-30
**/

import (
	"container/list"
	"errors"
	"fmt"
	"gopkg.in/redis.v3"
	"sync"
)

var reIDsClients *list.List
var reIDsMutex *sync.Mutex

var redisHost string
var redisPort int
var redisDb int64

var redisCount int

func init() {
	reIDsClients = list.New()
	reIDsMutex = new(sync.Mutex)
	redisCount = 4
}

//PopRedisClient 从redis连接池中获取一个redis连接,用完之后要调用PushRedisClient放回去
func PopRedisClient() *redis.Client {
	reIDsMutex.Lock()
	defer reIDsMutex.Unlock()
	//连接池不够用了
	if reIDsClients.Len() == 0 {
		for i := 0; i < redisCount; i++ {
			rdClient := redis.NewClient(&redis.Options{
				Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
				Password: "",      // no password set
				DB:       redisDb, // use default DB
			})
			if rdClient == nil {
				return nil
			}
			reIDsClients.PushBack(rdClient)
		}
		redisCount += redisCount
	}

	//Trace("redis queue size: %d, all: %d", reIDsClients.Len(), redisCount)
	client := reIDsClients.Front().Value.(*redis.Client)
	reIDsClients.Remove(reIDsClients.Front())
	return client
}

//ConnectRedis 连接到redis服务器,在程序启动时调用
func ConnectRedis(host string, port int, db int64) error {
	redisHost = host
	redisPort = port
	redisDb = db

	for i := 0; i < redisCount; i++ {
		rdClient := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
			Password: "",      // no password set
			DB:       redisDb, // use default DB
		})
		if rdClient == nil {
			return errors.New("connect redis fail")
		}
		reIDsClients.PushBack(rdClient)
	}
	return nil
}

//PushRedisClient 将一个redis连接放回reds连接池
func PushRedisClient(client *redis.Client) {
	reIDsMutex.Lock()
	defer reIDsMutex.Unlock()

	reIDsClients.PushBack(client)

	//空余太多连接
	if redisCount > 100 && reIDsClients.Len() >= (redisCount*3/4) {
		Info("too many redis connect, clear, free-connect-cnt: %d, total-connect-cnt: %d", reIDsClients.Len(), redisCount)
		count := redisCount / 2
		for i := 0; i < count && reIDsClients.Len() > 1; i++ {
			client := reIDsClients.Front().Value.(*redis.Client)
			reIDsClients.Remove(reIDsClients.Front())
			client.Close()
		}
		redisCount -= count
	}
}
