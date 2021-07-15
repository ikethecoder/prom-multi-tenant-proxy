package main

import (
	"log"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/kelseyhightower/envconfig"
	"github.com/ikethecoder/prom-multi-tenant-proxy/internal/app/query-proxy"
	"github.com/ikethecoder/prom-multi-tenant-proxy/internal/pkg"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const ns = "namespace"

func main() {
    var s pkg.Specification
    err := envconfig.Process("myapp", &s)
    if err != nil {
        log.Fatal(err.Error())
    }
		s.LCache = cache.New(10*time.Minute, 10*time.Minute)

    format := "Debug: %v\nPort: %d\nPrometheusUrl: %s\n"
    _, err = fmt.Printf(format, s.Debug, s.Port, s.PrometheusUrl)
    if err != nil {
        log.Fatal(err.Error())
    }

	proxy.Serve (&s)
	
}
