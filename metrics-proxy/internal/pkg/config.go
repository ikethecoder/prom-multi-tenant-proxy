package pkg

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// LabelMap contains a list of label to namespace mappings
type LabelNamespaceMap struct {
	MatchLabel string `yaml:"matchLabel"`
	NamespaceLabel string `yaml:"namespaceLabel"`
	Matches map[string]Namespace `yaml:"matches"`
}

// The namespace that the service maps to
type Namespace struct {
	Name  string `yaml:"namespace"`
}

// ParseConfig read a configuration file in the path `location` and returns an Authn object
func ParseConfig(location *string) (*LabelNamespaceMap, error) {
	data, err := ioutil.ReadFile(*location)
	if err != nil {
		return nil, err
	}
	labelMap := LabelNamespaceMap{}
	err = yaml.Unmarshal([]byte(data), &labelMap)
	if err != nil {
		return nil, err
	}
	return &labelMap, nil
}
