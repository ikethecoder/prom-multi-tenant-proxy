package proxy

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/ikethecoder/prom-multi-tenant-proxy/internal/pkg"
	"github.com/ikethecoder/prom-multi-tenant-proxy/pkg/injector"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// ReversePrometheus a
func ReversePrometheus(reverseProxy *httputil.ReverseProxy, prometheusServerURL *url.URL, config *pkg.Specification) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checkRequest(r, prometheusServerURL, config)
		reverseProxy.ServeHTTP(w, r)
		log.Printf("[TO]\t%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
	}
}

func modifyRequest(r *http.Request, prometheusServerURL *url.URL, prometheusQueryParameter string, config *pkg.Specification) error {
	namespaces := r.Context().Value(Namespace)
	expr, err := parser.ParseExpr(r.FormValue(prometheusQueryParameter))
	if err != nil {
		return err
	}

	err = injector.SetRecursive(expr, []*labels.Matcher{
		{
			Name:  config.NamespaceLabel,
			Type:  labels.MatchRegexp,
			Value: "^(" + strings.Join(namespaces.([]string), "|") + ")$",
		},
	})
	if err != nil {
		return err
	}

	if r.Method == "GET" {
		q := r.URL.Query()
		q.Set(prometheusQueryParameter, expr.String())
		r.URL.RawQuery = q.Encode()
		log.Println("TRANSFORMED QUERY TO", r.URL)
	} else {
		form := r.Form
		form.Set(prometheusQueryParameter, expr.String())
		body := form.Encode()
		r.Body = io.NopCloser(strings.NewReader(body))
		r.ContentLength = int64(len(body))
		r.URL.RawQuery = ""
		r.URL.Fragment = ""
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		log.Println("TRANSFORMED BODY TO", body)
	}

	return nil
}

func checkRequest(r *http.Request, prometheusServerURL *url.URL, config *pkg.Specification) error {
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
	r.Host = prometheusServerURL.Host
	r.URL.Scheme = prometheusServerURL.Scheme
	r.URL.Host = prometheusServerURL.Host
	r.Header.Set("X-Forwarded-Host", r.Host)
	return nil
}
