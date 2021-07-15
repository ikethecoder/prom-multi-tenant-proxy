package proxy

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/patrickmn/go-cache"
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

		// buf, err := json.MarshalIndent(tok, "", "  ")
		// if err != nil {
		//   fmt.Printf("failed to generate JSON: %s\n", err)
		// 	writeUnauthorisedResponse(w)
		//   return
		// }
		// fmt.Printf("%s\n", buf)
		
		claims := tok.PrivateClaims()
		log.Println("azp = ", claims["azp"])
		log.Println("usr = ", claims["preferred_username"])
		//fmt.Printf("sub\t%v\n", tok.Subject())
		// for key, value := range claims {
		// 	fmt.Printf("%s\t%v\n", key, value)
		// }

		var cacheKey string = claims["preferred_username"].(string)
		log.Println("CACHE KEY", cacheKey)

		if labels, found := config.LCache.Get(cacheKey); found {
			log.Println("CACHE HIT!", labels)
			ctx := context.WithValue(r.Context(), Namespace, labels.([]string))
			handler(w, r.WithContext(ctx))
		} else {

			client := &http.Client{}

			u, err := url.Parse(config.ResourceServerUrl)
			if err != nil {
				log.Fatal(err)
				writeUnauthorisedResponse(w)
				return
			}

			rr := http.Request{
				Method: "GET",
				URL:    u,
			}
			rr.Header = http.Header{
				"Authorization": []string{r.Header.Get("Authorization")},
			}
		
			resp, err := client.Do(&rr)
			if err != nil {
				http.Error(w, "Server Error", http.StatusInternalServerError)
				log.Println("ServeHTTP:", err)
				writeUnauthorisedResponse(w)
				return
			}
			defer resp.Body.Close()
		
			body, err := ioutil.ReadAll(resp.Body)

			var labels []string

			log.Println("AUTHZ", r.RemoteAddr, resp.Status)

			json.Unmarshal([]byte(body), &labels)

			log.Println("filter", labels)

			if (resp.StatusCode == 200) {
				log.Println("CACHING!", cacheKey, labels)
				config.LCache.Set(cacheKey, labels, cache.DefaultExpiration)
			}

			ctx := context.WithValue(r.Context(), Namespace, labels)
			handler(w, r.WithContext(ctx))
		}
	}
}

func writeUnauthorisedResponse(w http.ResponseWriter) {
	//w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	w.WriteHeader(401)
	w.Write([]byte("{\"status\":\"error\",\"error\":\"Blocked Access\"}\n"))
}
