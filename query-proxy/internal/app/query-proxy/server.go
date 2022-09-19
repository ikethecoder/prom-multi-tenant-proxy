package proxy

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/ikethecoder/prom-multi-tenant-proxy/internal/pkg"
)

// Serve serves
func Serve(config *pkg.Specification) error {
	//prometheusServerURL, _ := url.Parse(config.PrometheusUrl)
	serveAt := fmt.Sprintf(":%d", config.Port)
	// authConfigLocation := c.String("auth-config")
	// authConfig, _ := pkg.ParseConfig(&authConfigLocation)

	// initialize a reverse proxy and pass the actual backend server url here
	proxy, err := NewProxy(config.PrometheusUrl, config)
	if err != nil {
			panic(err)
	}

	// handle all requests to your server using the proxy
	http.HandleFunc("/", LogRequest(AuthStub(ProxyRequestHandler(proxy), config)))
	log.Fatal(http.ListenAndServe(serveAt, nil))


	// http.HandleFunc("/", createHandler(prometheusServerURL, config))
	// if err := http.ListenAndServe(serveAt, nil); err != nil {
	// 	log.Fatalf("Prometheus multi tenant proxy can not start %v", err)
	// 	return err
	// }
	return nil
}

func ProxyRequestHandler(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
	}
}

// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string, config *pkg.Specification) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
			return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
			originalDirector(req)
			modRequest(req, url, config)
	}

	//proxy.ModifyResponse = modifyResponse()
	proxy.ErrorHandler = errorHandler()
	return proxy, nil
}

func modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {

			return errors.New("response body is invalid")
	}
}

func errorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
			fmt.Printf("Got error while handling request: %v \n", err)
			return
	}
}

func modRequest(r *http.Request, prometheusServerURL *url.URL, config *pkg.Specification) error {
	log.Println("MOD")
	r.Host = prometheusServerURL.Host
	r.Header.Set("X-Forwarded-Host", r.Host)

	if r.URL.Path == "/api/v1/query" || r.URL.Path == "/api/v1/query_range" {
		if err := modifyRequest(r, prometheusServerURL, "query", config); err != nil {
			return err
		}
	}
	if r.URL.Path == "/api/v1/series" {
		if err := modifyRequest(r, prometheusServerURL, "match[]", config); err != nil {
			return err
		}
	}

	return nil
}

// func createHandler(prometheusServerURL *url.URL, config *pkg.Specification) http.HandlerFunc {
// 	reverseProxy := httputil.NewSingleHostReverseProxy(prometheusServerURL)
// 	// https://blog.joshsoftware.com/2021/05/25/simple-and-powerful-reverseproxy-in-go/
// 	originalDirector := reverseProxy.Director
// 	reverseProxy.Director = func(req *http.Request) {
// 			originalDirector(req)
// 			modRequest(req)
// 	}
// 	//return LogRequest(reverseProxy)
// 	return LogRequest(AuthStub(ReversePrometheus(reverseProxy, prometheusServerURL, config), config))
// }
