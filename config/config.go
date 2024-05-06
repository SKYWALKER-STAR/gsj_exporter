// Copyright 2022 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"os"
	"sync"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

var (
	configName = "config.yml"
)

type Config struct {
	Log       Log        `yaml:"log"`
	Instances []Instance `yaml:"instances"`
}

type Log struct {
	Level  string `yaml:"level"`
	MaxAge int    `yaml:"max_age"`
}

type Instance struct {
	InstanceId string `yaml:"instance_id"`
	ExcludeDbs string `yaml:"exclude_dbs"`
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	Db         string `yaml:"db"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
}

type Handler struct {
	sync.RWMutex
	Config *Config
}

func (ch *Handler) GetConfig() *Config {
	ch.RLock()
	defer ch.RUnlock()
	return ch.Config
}

func (ch *Handler) SetConfig(c *Config) {
	ch.RLock()
	defer ch.RUnlock()
	ch.Config = c
}

func (ch *Handler) ReloadConfig() error {
	config := &Config{}

	yamlReader, err := os.Open(configName)
	if err != nil {
		return fmt.Errorf("Error opening config file %q: %s", configName, err)
	}
	defer yamlReader.Close()
	decoder := yaml.NewDecoder(yamlReader)
	decoder.KnownFields(true)

	if err = decoder.Decode(config); err != nil {
		return fmt.Errorf("Error parsing config file %q: %s", configName, err)
	}

	ch.SetConfig(config)
	return nil
}

func (ch *Handler) WriteConfigFile(tgi Config) int {
	data,err := yaml.Marshal(tgi)
	if err != nil {
		panic(err)
		return -1
	}

	err = ioutil.WriteFile(configName,data,0777)
	if err != nil {
		panic(err)
		return -1
	}

	fmt.Println(tgi)

	return 0
}

func (ch *Handler) InitConfig(instances []Instance,log Log) Config {
	config := Config {
		Log: log,
		Instances: instances,
	}

	return config
}

func (ch *Handler) InitLog(level string,maxage int) Log {
	log := Log {
		Level: level,
		MaxAge: maxage,
	}

	return log
}

func (ch *Handler) InitInstance(instance_id string, excludedbs string, hosts string, port string, db string, user string, passwd string) Instance {
	instance := Instance {
		InstanceId: instance_id,
		ExcludeDbs: excludedbs,
		Host: hosts,
		Port: port,
		Db: db,
		User: user,
		Password: passwd,
	}

	return instance
}

func (ch *Handler) InitFromUrl(instance_id string, excludedbs string, hosts string, port string, db string, user string, passwd string,log_level string,log_max_age int) Config {

	log := ch.InitLog(log_level,log_max_age)
	instance := ch.InitInstance(instance_id,excludedbs,hosts,port,db,user,passwd)
	instances := []Instance{instance}

	config := ch.InitConfig(instances,log)

	return config
}
