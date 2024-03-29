// Copyright 2020 ikethecoder
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

package main

//	"encoding/json"

import (
	"flag"
	"fmt"
	"time"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/patrickmn/go-cache"
	"github.com/kelseyhightower/envconfig"
	"github.com/ikethecoder/prom-multi-tenant-proxy/internal/pkg"
)

var (
	lcache  *cache.Cache
)

// export MYAPP_NAMESPACELABEL=abc
type Specification struct {
    Debug          bool `default:false`
    Port           int `required:"true", default: 9092`
    MetricsUrl     string `required:"true"`
	KongUrl   string `required:"true", default:"http://kong:8001"`
}

func main() {
    var s Specification
    err := envconfig.Process("myapp", &s)
    if err != nil {
        log.Fatal(err.Error())
	}
	log.SetFormatter(&log.JSONFormatter{})
	if s.Debug {
		log.SetLevel(log.DebugLevel)
	}
	format := "Debug: %v\nPort: %d\nMetricsUrl: %s\nKongAdmin: %s\n"
	log.Debug(fmt.Sprintf(format, s.Debug, s.Port, s.MetricsUrl, s.KongUrl))

	flag.Parse()

	lcache := cache.New(1*time.Minute, 1*time.Minute)

	labelMap, err := pkg.ParseConfig(&s.KongUrl)
	if err != nil {
		log.Error(err.Error())
	} else {
		lcache.Set("kong-services", labelMap, cache.DefaultExpiration)
	}
	
	handler := &proxy{}
	handler.forwardUrl = s.MetricsUrl
	handler.kongUrl = s.KongUrl
	handler.lcache = lcache

	addr := fmt.Sprintf("0.0.0.0:%d", s.Port)

	log.Info("Starting proxy server on ", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Error("ListenAndServe:", err)
	}
}
