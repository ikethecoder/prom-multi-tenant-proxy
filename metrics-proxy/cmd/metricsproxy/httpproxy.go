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

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/patrickmn/go-cache"
	"github.com/ikethecoder/prom-multi-tenant-proxy/internal/app/prom"
	"github.com/ikethecoder/prom-multi-tenant-proxy/internal/pkg"
	dto "github.com/prometheus/client_model/go"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
	"Content-Length",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

type proxy struct {
	forwardUrl string
	kongUrl string
	lcache *cache.Cache
}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Debug("R=", req.RemoteAddr, " M=", req.Method, " U=", req.URL, " S=", req.URL.Scheme)

	// if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
	// 	msg := "unsupported protocal scheme " + req.URL.Scheme
	// 	http.Error(wr, msg, http.StatusBadRequest)
	// 	log.Println(msg)
	// 	return
	// }

	client := &http.Client{}

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	req.RequestURI = ""

	u, err := url.Parse(p.forwardUrl)
	if err != nil {
		log.Fatal(err)
	}
	u.Path = req.URL.Path

	if req.URL.Path != "/metrics" {
		http.Error(wr, "Invalid path", http.StatusNotFound)
		return
	}

	// Construct request to send to origin server
	rr := http.Request{
		Method: req.Method,
		URL:    u,
		Header: req.Header,
		Body:   req.Body,
		// TODO: Is this correct for a 0 value?
		//       Perhaps a 0 may need to be reinterpreted as -1?
		ContentLength: req.ContentLength,
		Close:         req.Close,
	}

	delHopHeaders(req.Header)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	resp, err := client.Do(&rr)
	if err != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		log.Error("ServeHTTP:", err)
		return
	}
	defer resp.Body.Close()

	log.Debug(req.RemoteAddr, " ", resp.Status)

	delHopHeaders(resp.Header)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)

	mfChan := make(chan *dto.MetricFamily, 1024)
	go func() {
		if err := prom.ParseReader(resp.Body, mfChan); err != nil {
			log.Error("ERROR reading metrics:", err)
			return
		}
	}()

	if labelMap, found := p.lcache.Get("kong-services"); found {
		log.Debug("Found Kong config in cache.. using it!")
		prom.Write(mfChan, wr, *labelMap.(*pkg.LabelNamespaceMap))
	} else {
		log.Debug("Refreshing kong services...")
		labelMap, err := pkg.ParseConfig(&p.kongUrl)
		if err != nil {
			log.Error(err.Error())
			http.Error(wr, "Server Error", http.StatusInternalServerError)
		}
		p.lcache.Set("kong-services", labelMap, cache.DefaultExpiration)
		prom.Write(mfChan, wr, *labelMap)
	}
}
