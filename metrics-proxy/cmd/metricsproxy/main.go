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
	"log"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"github.com/ikethecoder/prom-multi-tenant-proxy/internal/pkg"
)

// export MYAPP_NAMESPACELABEL=abc
type Specification struct {
    Debug          bool `default:false`
    Port           int `required:"true", default: 9092`
    MetricsUrl     string `required:"true"`
	LabelMapPath   string `required:"true", default:"labelmap.yml"`
}

func main() {
    var s Specification
    err := envconfig.Process("myapp", &s)
    if err != nil {
        log.Fatal(err.Error())
    }
    format := "Debug: %v\nPort: %d\nMetricsUrl: %s\n"
    _, err = fmt.Printf(format, s.Debug, s.Port, s.MetricsUrl)
    if err != nil {
        log.Fatal(err.Error())
    }

	flag.Parse()

	labelMap, err := pkg.ParseConfig(&s.LabelMapPath)
    if err != nil {
        log.Fatal(err.Error())
    }
	
	handler := &proxy{}
	handler.forwardUrl = s.MetricsUrl
	handler.labelMap = *labelMap

	addr := fmt.Sprintf("0.0.0.0:%d", s.Port)

	log.Println("Starting proxy server on", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Println("ListenAndServe:", err)
	}
}
