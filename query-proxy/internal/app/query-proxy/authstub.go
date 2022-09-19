package proxy

import (
	"context"
	// "fmt"
	// "io/ioutil"
	"encoding/json"
	"log"
	"net/http"
	// "net/url"
	// "strings"

	// "github.com/patrickmn/go-cache"
	// "github.com/lestrrat-go/jwx/jwk"
	// "github.com/lestrrat-go/jwx/jwt"
	
	"github.com/ikethecoder/prom-multi-tenant-proxy/internal/pkg"
)



// JWTAuth can be used as a middleware chain to authenticate users before proxying a request
func AuthStub(handler http.HandlerFunc, config *pkg.Specification) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var labels []string

		var body = "[]"
		json.Unmarshal([]byte(body), &labels)

		log.Println("filter", labels)


		ctx := context.WithValue(r.Context(), Namespace, labels)
		handler(w, r.WithContext(ctx))
	}
}
