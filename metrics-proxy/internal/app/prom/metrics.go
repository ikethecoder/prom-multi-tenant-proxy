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

package prom

import (
//	"bufio"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/common/expfmt"
	"github.com/ikethecoder/prom-multi-tenant-proxy/internal/pkg"
	dto "github.com/prometheus/client_model/go"
)

// ParseReader consumes an io.Reader and pushes it to the MetricFamily
// channel. It returns when all MetricFamilies are parsed and put on the
// channel.
func ParseReader(in io.Reader, ch chan<- *dto.MetricFamily) error {
	defer close(ch)
	// We could do further content-type checks here, but the
	// fallback for now will anyway be the text format
	// version 0.0.4, so just go for it and see if it works.
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(in)
	if err != nil {
		return fmt.Errorf("reading text format failed: %v", err)
	}
	for _, mf := range metricFamilies {
		ch <- mf
	}
	return nil
}

// AddLabel allows to add key/value labels to an already existing Family.
func AddLabel(item *dto.Metric, name string, val string) {
	item.Label = append(item.Label, &dto.LabelPair{
		Name:  &name,
		Value: &val,
	})
}

func Write(ch chan *dto.MetricFamily, ioWriter io.Writer, labelMap pkg.LabelNamespaceMap) {
	w := ioWriter

	for mf := range ch {
		for _, metric := range mf.Metric {
			for _, label := range metric.Label {
				if *label.Name == labelMap.MatchLabel {
					mappedValue, ok := labelMap.Matches[*label.Value]
					if ok {
						AddLabel(metric, labelMap.NamespaceLabel, mappedValue.Name)
					}
					break
				}
			}
		}
		_, err := expfmt.MetricFamilyToText(w, mf)
		if err != nil {
			log.Error(err)
		}
	}

	//w.Flush()
}
