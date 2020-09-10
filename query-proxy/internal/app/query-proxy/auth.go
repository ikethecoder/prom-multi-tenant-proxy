package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat/go-jwx/jwk"
	
	"github.com/ikethecoder/prom-multi-tenant-proxy/internal/pkg"
)

type key int

const (
	//Namespace Key used to pass prometheus tenant id though the middleware context
	Namespace key = iota
	realm         = "Prometheus multi-tenant proxy"
)

const jwksURL = "https://auth-qwzrwc-dev.pathfinder.gov.bc.ca/auth/realms/gwa/protocol/openid-connect/certs"

func getKey(token *jwt.Token) (interface{}, error) {

    // TODO: cache response so we don't have to make a request every time 
    // we want to verify a JWT
    set, err := jwk.FetchHTTP(jwksURL)
    if err != nil {
        return nil, err
    }

    keyID, ok := token.Header["kid"].(string)
    if !ok {
        return nil, errors.New("expecting JWT header to have string kid")
    }

    if key := set.LookupKeyID(keyID); len(key) == 1 {
        return key[0].Materialize()
    }

    return nil, errors.New("unable to find key")
}


// JWTAuth can be used as a middleware chain to authenticate users before proxying a request
func JWTAuth(handler http.HandlerFunc, authConfig *pkg.Authn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var token string
		var namespace string

        // Get token from the Authorization header
        // format: Authorization: Bearer 
		tokens, ok := r.Header["Authorization"]
		if !ok {
			writeUnauthorisedResponse(w)
			return
		}
        if ok && len(tokens) >= 1 {
            token = tokens[0]
            token = strings.TrimPrefix(token, "Bearer ")
        }

        // If the token is empty...
        if token == "" {
            // If we get here, the required token is missing
            http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
            return
        }

		tok, err := jwt.Parse(token, getKey)
		if err != nil {
			panic(err)
		}
		claims := tok.Claims.(jwt.MapClaims)
		for key, value := range claims {
			fmt.Printf("%s\t%v\n", key, value)
		}

		// Use the 'team' claim for the namespace
		namespace = fmt.Sprintf("%v", claims["team"])
		if !ok {
			writeUnauthorisedResponse(w)
			return
		}
		ctx := context.WithValue(r.Context(), Namespace, namespace)
		handler(w, r.WithContext(ctx))
	}
}

func writeUnauthorisedResponse(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	w.WriteHeader(401)
	w.Write([]byte("Unauthorised\n"))
}
