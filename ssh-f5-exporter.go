package main

import (
	"net/http"
	"github.com/xiachongbuyubing/ssh-f5-exporter/util"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//Define a struct for you collector that contains pointers
//to prometheus descriptors for each metric you wish to expose.
//Note you can also include fields of other types if they provide utility
//but we just won't be exposing them as metrics.
type serviceCollector struct {
	serviceMetric *prometheus.Desc
	host          string
	cmd		  string
}

//You must create a constructor for you collector that
//initializes every descriptor and returns a pointer to the collector
func newserviceCollector(host string, cmd string) *serviceCollector {
	return &serviceCollector{
		serviceMetric: prometheus.NewDesc("service_status",
			"Shows a service of f5 current status ",
			[]string{"service_name", "status_string"}, nil,
		),
		host: host,
		cmd: cmd,
	}
}

//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (collector *serviceCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.serviceMetric
}

//Collect implements required collect function for all promehteus collectors
func (collector *serviceCollector) Collect(ch chan<- prometheus.Metric) {

	//Implement logic here to determine proper metric value to return to prometheus
	//for each descriptor or call other functions that do so.
	var markValue float64
	var nameValue []string
	var statusValue []string

	servicelist, err := util.Excutescript(collector.host, "22", "root", "default", collector.cmd)
	if err != nil {
		log.Error(err)
		return
	}
	nameValue, statusValue = util.Dealstr(servicelist)

	if err != nil {
		log.Error(err)
	}

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	for i := 0; i < len(nameValue); i++ {
		markValue = 0
		if strings.Contains(statusValue[i], "run") {
			markValue = 1
		}
		ch <- prometheus.MustNewConstMetric(collector.serviceMetric, prometheus.GaugeValue, markValue, nameValue[i], statusValue[i])
	}

}
//handler handle a host from prometheus configuration and register
func handler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	script := r.URL.Query().Get("script")
	service := newserviceCollector(target, script)
	registry := prometheus.NewRegistry()
	registry.MustRegister(service)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {

	//This section will start the HTTP server and expose
	//any metrics on the /ssh endpoint.
	http.HandleFunc("/ssh", handler)
	log.Info("Beginning to serve on port :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
