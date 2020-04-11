package main

import (
	"os"
	"fmt"
	"regexp"
	"time"

	"github.com/go-redis/redis/v7"

	log "github.com/sirupsen/logrus"
)

// struct that contains pointer to information about game state
type RedisConnection struct {
	gameConnections *map[string]*Hub
}

const (
	keyspaceFormat = "__keyspace@0__:%s"
	keyspaceEvents = "K$"
)

var (
	redisRegex = regexp.MustCompile("redis://rediscloud:(.*)@(.*)")
	match = redisRegex.FindStringSubmatch(os.Getenv("REDISCLOUD_URL"))

	// construct redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr:     match[2],
		Password: match[1],
		DB:       0,  // use default DB
	})

	// empty subscription for initialization
	pubsub = redisClient.Subscribe()

	// regular expression to parse key from pubsub messages
	keyspaceRegex = regexp.MustCompile(fmt.Sprintf(keyspaceFormat, "(.*)"))
)

// simple wrapper to fetch all keys from redis
func (rc *RedisConnection) GetKeys() (map[string]bool) {
	val := redisClient.Do("KEYS", "*").Val().([]interface{})
	keys := make(map[string]bool)
	for i := 0; i < len(val); i++ {
		keys[val[i].(string)] = true
	}

	return keys
}

// simple wrapper function around fetching from redis
func (rc *RedisConnection) GetKey(key string) (string, error) {
	return redisClient.Get(key).Result()
}

// simple wrapper function around setting to redis
func (rc *RedisConnection) SetKey(key string, value []byte) error {
	return redisClient.Set(key, value, 1 * time.Hour).Err()
}

// subscribe to a given key channel
func (rc *RedisConnection) Subscribe(key string) error {
	log.Infof("subscribing to key %s", key)
	return pubsub.Subscribe(fmt.Sprintf(keyspaceFormat, key))
}


// enable keyspace events
func (rc *RedisConnection) SetKeyspaceEvents() error {
	log.Infof("setting keyspace events to %s", keyspaceEvents)
	return redisClient.ConfigSet("notify-keyspace-events", keyspaceEvents).Err()
}

// unsubscribe to a given key channel
func (rc *RedisConnection) Unsubscribe(key string) error {
	log.Infof("unsubscribing from key %s", key)
	return pubsub.Unsubscribe(fmt.Sprintf(keyspaceFormat, key))
}

// propagates updates to all sockets for any key change
func (rc *RedisConnection) PropagateUpdate() {
	defer func() {
		// gracefully close pubsub connection
		pubsub.Close()
	}()
	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			log.Errorf("error receiving message %v", err)
			return
		}

		log.Infof("message %s", msg)
		match := keyspaceRegex.FindStringSubmatch(msg.Channel)
		identifier := match[1]

		log.Infof("received update for identifier %s", identifier)
		value, err := rc.GetKey(identifier)
		(*rc.gameConnections)[identifier].inbound <- []byte(value)
	}
}
