package pkg

import (
	"log"
	"fmt"
	"strings"
	// "io/ioutil"

	"encoding/json"
	"net/http"
)

// LabelMap contains a list of label to namespace mappings
type LabelNamespaceMap struct {
	MatchLabel string
	NamespaceLabel string
	Matches map[string]Namespace
}

// The namespace that the service maps to
type Namespace struct {
	Name  string
}

type KongServices struct {
    Next string `json:"next"`
	Services []KongService `json:"data"`
}

type KongService struct {
	Name string `json:"name"`
	Tags []string `json:"tags"`
}

func ParseConfig(location *string) (*LabelNamespaceMap, error) {
	labelMap := LabelNamespaceMap{}
	labelMap.MatchLabel = "service"
	labelMap.NamespaceLabel = "namespace"
	labelMap.Matches = make(map[string]Namespace)
	
	var nextUrl string

	nextUrl = "/services"

	for nextUrl != "" {
		var data KongServices

		log.Println("Getting", nextUrl)
		resp, err := http.Get(fmt.Sprintf("%s%s", *location, nextUrl))
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		defer resp.Body.Close()

		json.NewDecoder(resp.Body).Decode(&data)
		log.Println("Decoded - is next?", data.Next, len(data.Services))

		nextUrl = data.Next

		for _, svc := range data.Services {

			// Find the tag that starts with ns.*
			for _, tag := range svc.Tags {
				if strings.HasPrefix (tag, "ns.") {
					var nsTag = strings.TrimPrefix (tag, "ns.")
					log.Println("Service=", svc.Name, "Namespace=", nsTag)

					match := Namespace{}
					match.Name = nsTag
					labelMap.Matches[svc.Name] = match
					break
				}
			}
		}
	}
	return &labelMap, nil
}
