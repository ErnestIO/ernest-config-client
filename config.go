/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package ernest_config_client

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/nats-io/nats"
	"gopkg.in/redis.v3"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var err error
var cfg map[string]interface{}

type config struct {
	uri      string
	nats     *nats.Conn
	postgres *gorm.DB
	redis    *redis.Client
}

func NewConfig(natsURI string) config {
	c := config{uri: natsURI}
	c.setup()

	return c
}

func (c *config) setup() {
	c.Nats()
}

func (c *config) Nats() *nats.Conn {
	if c.nats != nil {
		return c.nats
	}

	for c.nats == nil {
		c.nats, err = nats.Connect(c.uri)
		if err != nil {
			log.Println("Waiting for nats on " + c.uri + ". Retrying in 2 seconds ...")
			time.Sleep(time.Second * 2)
		}
	}
	return c.nats
}

func (c *config) Postgres(table string) *gorm.DB {
	var resp *nats.Msg

	for c.postgres == nil {
		resp, err = c.nats.Request("config.get.postgres", nil, time.Second)
		if err != nil {
			log.Println("Waiting for config.get.postgres response. Retrying in 5 seconds ...")
			time.Sleep(time.Second * 5)
			continue
		}

		err = json.Unmarshal(resp.Data, &cfg)
		if err != nil {
			log.Println("Invalid config.get.postgres response, received '" + string(resp.Data) + "'. Retrying in 5 seconds ...")
			time.Sleep(time.Second * 5)
			continue
		}

		uri := fmt.Sprintf("%s/%s?sslmode=disable", cfg["url"], table)
		c.postgres, err = gorm.Open("postgres", uri)
		if err != nil {
			log.Println("Unsuccesful connection to postgres ''. Retrying in 10 seconds ...")
			time.Sleep(time.Second * 10)
			continue
		}
	}

	return c.postgres
}

func (c *config) Redis() *redis.Client {
	if c.redis != nil {
		return c.redis
	}

	var redisCfg struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		DB       int64  `json:"db"`
	}

	for c.redis == nil {
		resp, err := c.nats.Request("config.get.redis", nil, time.Second)
		if err != nil {
			log.Println("Waiting for config.get.redis response. Retrying in 5 seconds ...")
			time.Sleep(time.Second * 5)
			continue
		}

		err = json.Unmarshal(resp.Data, &redisCfg)
		if err != nil {
			log.Println("Invalid config.get.redis response, received '" + string(resp.Data) + "'. Retrying in 5 seconds ...")
			time.Sleep(time.Second * 5)
			continue
		}

		redis := redis.NewClient(&redis.Options{
			Addr:     redisCfg.Addr,
			Password: redisCfg.Password,
			DB:       redisCfg.DB,
		})

		pong, err := redis.Ping().Result()
		if err != nil {
			log.Println("Redis responded with '" + pong + "' to a ping request. Reconnecting in 5 seconds ...")
			time.Sleep(time.Second * 5)
			continue
		}
		c.redis = redis
	}

	return c.redis
}
