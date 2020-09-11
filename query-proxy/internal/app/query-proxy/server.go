package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/ikethecoder/prom-multi-tenant-proxy/internal/pkg"
)

// Serve serves
func Serve(config *pkg.Specification) error {
	prometheusServerURL, _ := url.Parse(config.PrometheusUrl)
	serveAt := fmt.Sprintf(":%d", config.Port)
	// authConfigLocation := c.String("auth-config")
	// authConfig, _ := pkg.ParseConfig(&authConfigLocation)

	http.HandleFunc("/", createHandler(prometheusServerURL, config))
	if err := http.ListenAndServe(serveAt, nil); err != nil {
		log.Fatalf("Prometheus multi tenant proxy can not start %v", err)
		return err
	}
	return nil
}

func createHandler(prometheusServerURL *url.URL, config *pkg.Specification) http.HandlerFunc {
	reverseProxy := httputil.NewSingleHostReverseProxy(prometheusServerURL)
	return LogRequest(JWTAuth(ReversePrometheus(reverseProxy, prometheusServerURL, config), config))
}
