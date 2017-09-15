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
	"github.com/r3labs/akira"
	"gopkg.in/redis.v3"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var err error
var cfg map[string]interface{}

type Config struct {
	uri      string
	nats     akira.Connector
	postgres *gorm.DB
	redis    *redis.Client
}

func NewConfig(natsURI string) *Config {
	c := Config{uri: natsURI}
	c.setup()

	return &c
}

func (c *Config) setup() {
	c.Nats()
}

func (c *Config) SetConnector(conn akira.Connector) {
	c.nats = conn
}

func (c *Config) Nats() *nats.Conn {
	if c.nats != nil {
		return c.nats.(*nats.Conn)
	}

	for c.nats == nil {
		c.nats, err = nats.Connect(c.uri)
		if err != nil {
			log.Println("Waiting for nats on " + c.uri + ". Retrying in 2 seconds ...")
			time.Sleep(time.Second * 2)
			continue
		}
		log.Println("Successfully connected to nats on '" + c.uri + "'")
	}
	return c.nats.(*nats.Conn)
}

func (c *Config) Postgres(table string) *gorm.DB {
	var resp *nats.Msg
	var pgCfg map[string]interface{}

	for c.postgres == nil {
		resp, err = c.nats.Request("config.get.postgres", nil, time.Second)
		if err != nil {
			log.Println("Waiting for config.get.postgres response. Retrying in 5 seconds ...")
			time.Sleep(time.Second * 5)
			continue
		}

		err = json.Unmarshal(resp.Data, &pgCfg)
		if err != nil {
			log.Println("Invalid config.get.postgres response, received '" + string(resp.Data) + "'. Retrying in 5 seconds ...")
			time.Sleep(time.Second * 5)
			continue
		}

		uri := fmt.Sprintf("%s/%s?sslmode=disable", pgCfg["url"], table)
		c.postgres, err = gorm.Open("postgres", uri)
		if err != nil {
			log.Println("Unsuccesful connection to postgres '" + uri + "'. Retrying in 10 seconds ...")
			time.Sleep(time.Second * 10)
			c.postgres = nil
			continue
		}
		log.Println("Successfully connected to postgres on '" + uri + "'")
	}

	return c.postgres
}

func (c *Config) Redis() *redis.Client {
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
		log.Println("Successfully connected to redis on '" + redisCfg.Addr + "'")
		c.redis = redis
	}

	return c.redis
}

func (c *Config) GetConfig(ctype string, result interface{}) error {
	var msg *nats.Msg

	for msg == nil {
		msg, err = c.nats.Request("config.get."+ctype, nil, time.Second)
		if err != nil {
			log.Printf("Waiting for config.get.%s response. Retrying in 5 seconds ...", ctype)
			time.Sleep(time.Second * 5)
			continue
		}
	}

	return json.Unmarshal(msg.Data, result)
}
