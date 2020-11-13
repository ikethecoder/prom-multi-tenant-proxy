package pkg

import (
	"errors"
	"fmt"
	"strings"
	// "io/ioutil"

	log "github.com/sirupsen/logrus"

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

		log.Debug("Getting", nextUrl)
		resp, err := http.Get(fmt.Sprintf("%s%s", *location, nextUrl))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, errors.New("failed to get Kong Service list")
		}
		json.NewDecoder(resp.Body).Decode(&data)
		log.Debug("Decoded - is next?", data.Next, len(data.Services))

		nextUrl = data.Next

		for _, svc := range data.Services {

			// Find the tag that starts with ns.*
			for _, tag := range svc.Tags {
				if strings.HasPrefix (tag, "ns.") {
					var nsTag = strings.TrimPrefix (tag, "ns.")
					log.Debug("Service=", svc.Name, " Namespace=", nsTag)

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
