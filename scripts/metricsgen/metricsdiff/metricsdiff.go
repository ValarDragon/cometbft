// metricsdiff is a tool for generating a diff between two different files containing
// prometheus metrics. metricsdiff outputs which metrics have been added, removed,
// or have different sets of labels between the two files.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: %[1]s <path> <path>

Generate the diff between the two files of Prometheus metrics.
The input should have the format output by a Prometheus HTTP endpoint.
The tool indicates which metrics have been added, removed, or use different
label sets from path1 to path2.

`, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

// Diff contains the set of metrics that were modified between two files
// containing prometheus metrics output.
type Diff struct {
	Adds    []string
	Removes []string

	LabelDiffs []LabelDiff
}

// LabelDiff describes the label changes between two versions of the same metric.
type LabelDiff struct {
	MetricsName string
	Adds        []string
	Removes     []string
}

func main() {
	flag.Parse()
	if len(os.Args) != 3 {
		log.Fatalf("Incorrect number of arguments, expected 2 but received %d", len(os.Args)-1)
	}
	fa, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Unable to open file %s: %v", os.Args[1], err)
	}
	fb, err := os.Open(os.Args[2])
	if err != nil {
		log.Fatalf("Unable to open file %s: %v", os.Args[2], err)
	}
	md, err := MetricsDiffFromReaders(fa, fb)
	if err != nil {
		log.Fatalf("Generating diff: %v", err)
	}
	fmt.Printf("%s", md)
}

// MetricsDiffFromReaders parses the metrics present in the readers a and b and
// determines which metrics were added and removed in b.
func MetricsDiffFromReaders(a, b io.Reader) (MetricsDiff, error) {
	parser := expfmt.TextParser{}
	amf, err := parser.TextToMetricFamilies(a)
	if err != nil {
		return MetricsDiff{}, err
	}
	bmf, err := parser.TextToMetricFamilies(b)
	if err != nil {
		return MetricsDiff{}, err
	}

	md := MetricsDiff{}
	for name, afamily := range amf {
		bfamily, ok := bmf[name]
		if !ok {
			md.Removes = append(md.Removes, name)
			continue
		}

		labelsDiff := false
		aLabelSet := toSet(afamily.Metric[0].Label)
		bLabelSet := toSet(bfamily.Metric[0].Label)
		ld := LabelDiff{
			MetricsName: name,
		}
		for name := range aLabelSet {
			_, ok := bLabelSet[name]
			if !ok {
				ld.Removes = append(ld.Removes, name)
				labelsDiff = true
			}
		}
		for name := range bLabelSet {
			_, ok := aLabelSet[name]
			if !ok {
				ld.Adds = append(ld.Adds, name)
				labelsDiff = true
			}
		}
		if labelsDiff {
			md.LabelDiffs = append(md.LabelDiffs, ld)
		}
	}
	for name := range bmf {
		if _, ok := amf[name]; !ok {
			md.Adds = append(md.Adds, name)
		}
	}
	return md, nil
}

func toSet(lps []*dto.LabelPair) map[string]struct{} {
	m := make(map[string]struct{}, len(lps))
	for _, pair := range lps {
		m[*pair.Name] = struct{}{}
	}
	return m
}

func (m MetricsDiff) String() string {
	var s string
	if len(m.Adds) > 0 {
		s += "Adds: \n"
		for _, add := range m.Adds {
			s += fmt.Sprintf("+++ %s\n", add)
		}
	}
	if len(m.Removes) > 0 {
		s += "Removes: \n"
		for _, remove := range m.Removes {
			s += fmt.Sprintf("--- %s\n", remove)
		}
	}
	if len(m.LabelDiffs) > 0 {
		s += "Label Changes: \n"
		for _, ld := range m.LabelDiffs {
			s += fmt.Sprintf("Label: %s\n", ld.MetricsName)
			for _, add := range ld.Adds {
				s += fmt.Sprintf("+++ %s", add)
			}
			for _, remove := range ld.Removes {
				s += fmt.Sprintf("--- %s", remove)
			}
		}
	}
	return s
}
