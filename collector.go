package main

import (
	"log"
	"time"

	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

/*

	Allgemeine Tips:
		* Der Typ Exporter und seine Methoden würde ich der Übersicht halber gruppieren. Typische Anordnung: Typ, Konstruktor, Methoden
		* Das Parsen würde ich noch ein bisschen auslagern und ein paar Type einführen. Type wie [][]string sagen sehr wenig aus.

*/

type metric struct {
	metricsname string
	value       float64
	unit        string
	addr        string
}

// Exporter implements the prometheus.Collector interface. It exposes the metrics
// of a ipmi node.
type Exporter struct {
	IpmiBinary   string
	metrics      map[string]*prometheus.GaugeVec
	duration     prometheus.Gauge
	totalScrapes prometheus.Counter
	namespace    string
	replacer     *strings.Replacer
}

// NewExporter instantiates a new ipmi Exporter.
func NewExporter(ipmiBinary string, replacer *strings.Replacer) *Exporter {
	e := Exporter{
		IpmiBinary: ipmiBinary,
		namespace:  "ipmi",
		replacer:   replacer,
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: *namespace,
			Name:      "exporter_last_scrape_duration_seconds",
			Help:      "The last scrape duration.",
		}),
	}

	e.metrics = map[string]*prometheus.GaugeVec{}

	e.collect()
	return &e
}

func executeCommand(cmd string) (string, error) {
	parts := strings.Fields(cmd)
	out, err := exec.Command(parts[0], parts[1]).Output()
	return string(out), err
}

func createMetrics(e *Exporter, metric []metric) {
	for _, m := range metric {
		e.metrics[m.metricsname] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ipmi",
			Name:      m.metricsname,
			Help:      m.metricsname,
			ConstLabels: map[string]string{
				"unit": m.unit,
			},
		}, []string{"addr"})

		labels := prometheus.Labels{"addr": "localhost"}
		e.metrics[m.metricsname].With(labels).Set(m.value)
	}
}

// Describe Describes all the registered stats metrics from the ipmi node.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.metrics {
		m.Describe(ch)
	}

	ch <- e.duration.Desc()
}

// Collect collects all the registered stats metrics from the ipmi node.
func (e *Exporter) Collect(metrics chan<- prometheus.Metric) {
	e.collect()
	for _, m := range e.metrics {
		m.Collect(metrics)
	}

	metrics <- e.duration
}

func (e *Exporter) collect() {
	now := time.Now().UnixNano()

	output, err := executeCommand(e.IpmiBinary)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	splitted, err := splitAoutput(string(output))
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	convertedOutput, err := convertOutput(splitted, e.replacer)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	createMetrics(e, convertedOutput)

	e.duration.Set(float64(time.Now().UnixNano()-now) / 1000000000)

	if err != nil {
		log.Printf("could not retrieve ipmi metrics: %v", err)
	}
}
