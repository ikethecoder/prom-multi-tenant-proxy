package proxy

import (
	"context"
	"errors"
	"fmt"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	// "github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	// "github.com/lestrrat-go/jwx/jwa"
	
	"github.com/ikethecoder/prom-multi-tenant-proxy/internal/pkg"
)

type key int

const (
	//Namespace Key used to pass prometheus tenant id though the middleware context
	Namespace key = iota
	realm         = "Prometheus multi-tenant proxy"
)

func getKeySet(config *pkg.Specification) (*jwk.Set, error) {

    // TODO: cache response so we don't have to make a request every time 
    // we want to verify a JWT
    set, err := jwk.FetchHTTP(config.JwksUrl)
    if err != nil {
        return nil, err
    }

	return set, nil
}

func ParseToken(token *string, config *pkg.Specification) (jwt.Token, error)  {
	if config.VerifyToken {
		keySet, err := getKeySet(config)
		if err != nil {
			panic(err)
		}
		tok, err := jwt.Parse(strings.NewReader(*token), jwt.WithKeySet(keySet))
		if err != nil {
			return nil, errors.New(fmt.Sprintf("JWT validation failed - %v", err))
		}
		return tok, nil
	} else {
		tok, err := jwt.Parse(strings.NewReader(*token))
		if err != nil {
			return nil, err
		}
		return tok, nil
	}

}

// JWTAuth can be used as a middleware chain to authenticate users before proxying a request
func JWTAuth(handler http.HandlerFunc, config *pkg.Specification) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var token string
		var namespace string

		log.Println("Verify Token:", len(r.Header["Authorization"]))

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

		// var tok jwt.Token
		tok, err := ParseToken(&token, config)
		if err != nil {
			log.Println(err)
			writeUnauthorisedResponse(w)
			return
		}

		// jwt.WithVerify(jwa.RS256, &privKey.PublicKey)
		// if config.VerifyToken {
		// 	//key, _ := getKey(token, config)
		// 	// , jwt.WithVerify(jwa.RS256, key)
		// 	tok, err := jwt.Parse(strings.NewReader(token))
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// } else {
		// 	tok, err := jwt.Parse(strings.NewReader(token))
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// }

		buf, err := json.MarshalIndent(tok, "", "  ")
		if err != nil {
		  fmt.Printf("failed to generate JSON: %s\n", err)
		  return
		}
		fmt.Printf("%s\n", buf)
		
		claims := tok.PrivateClaims()
		for key, value := range claims {
			fmt.Printf("%s\t%v\n", key, value)
		}

		namespace = fmt.Sprintf("%v", claims[config.NamespaceClaim])

		ctx := context.WithValue(r.Context(), Namespace, namespace)
		handler(w, r.WithContext(ctx))
	}
}

func writeUnauthorisedResponse(w http.ResponseWriter) {
	//w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	w.WriteHeader(400)
	w.Write([]byte("{\"status\":\"error\",\"error\":\"Blocked Access\"}\n"))
}
