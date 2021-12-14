package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"text/template"

	"github.com/go-chi/chi"
)

const (
	host                = "127.0.0.1"
	port                = "8080"
	metric_type_gauge   = "gauge"
	metric_type_counter = "counter"
)

type gauge struct {
	v float64
}

type counter struct {
	v int64
}

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

var gmetrics = make(map[string]gauge)
var cmetrics = make(map[string]counter)
var templateDataMap = make(map[string]interface{})
var metrics = make(map[string]Metric)

func main() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go runServer()
	<-signalChannel
	endServer()
}

//next step func gaugeHndlr(w http.ResponseWriter, r *http.Request) {
func jsonUpdMetricsHandler(w http.ResponseWriter, r *http.Request) {
	m, err := GetJRecBody(r)
	if err != nil {
		http.Error(w, "Internal error during JSON parsing", http.StatusInternalServerError)
		return
	}
	switch m.MType {
	case metric_type_counter:
		if metrics[m.ID].Delta == nil {
			metrics[m.ID] = *m
		} else {
			*metrics[m.ID].Delta += *m.Delta
		}
	case metric_type_gauge:
		metrics[m.ID] = *m
	default:
		log.Printf("Metric type '%s' is not expected. Skipping.", m.MType)
	}
	w.WriteHeader(http.StatusOK)
	r.Body.Close()

}

func gaugeHndlr(w http.ResponseWriter, r *http.Request) {
	mName, mValue := getterReqBody(r)
	value, err := strconv.ParseFloat(string(mValue), 64)
	if err != nil {
		http.Error(w, "parsing error. Bad request", http.StatusBadRequest)
		return
	}
	gmetrics[mName] = gauge{v: value}
	//log.Println("gauga updated")
	w.WriteHeader(http.StatusOK)
}

func jsonGetterMetricHandler(w http.ResponseWriter, r *http.Request) {
	m, err := GetJRecBody(r)
	r.Body.Close()
	if err != nil {
		http.Error(w, "Internal error during JSON unmarshal", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	data, found := metrics[m.ID]
	if !found {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	res, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Internal error during JSON marshal", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	//w.Write(res)
	w.Write([]byte(fmt.Sprint(res)))
}

func counterHndlr(w http.ResponseWriter, r *http.Request) {
	mName, mVal := getterReqBody(r)
	value, err := strconv.ParseInt(string(mVal), 10, 64)
	if err != nil {
		http.Error(w, "parsing error", http.StatusBadRequest)
		return
	}
	cmetrics[mName] = counter{cmetrics[mName].v + value}
	//log.Println("caunta updated")
	w.WriteHeader(http.StatusOK)
}

func getterMetricHandler(w http.ResponseWriter, r *http.Request) {
	metricType := strings.Split(r.URL.Path, "/")[2]
	metricName := strings.Split(r.URL.Path, "/")[3]
	if metricType == "counter" {
		if val, found := cmetrics[metricName]; found {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprint(val.v)))
		} else {
			http.Error(w, "There is no metric you requested", http.StatusNotFound)
		}
	}
	if metricType == "gauge" {
		if val, found := gmetrics[metricName]; found {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprint(val.v)))
		} else {
			http.Error(w, "There is no metric you requested", http.StatusNotFound)
		}
	}
}
func getterAllHndlr(w http.ResponseWriter, r *http.Request) {
	templateDataMap["Gmetrics"] = gmetrics
	templateDataMap["Cmetrics"] = cmetrics
	htmlPage, err := os.ReadFile("./allmetrics.html")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	tmpl := template.Must(template.New("").Parse(string(htmlPage)))
	tmpl.Execute(w, templateDataMap)
}

//next step
func GetJRecBody(r *http.Request) (*Metric, error) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}
	m := &Metric{}
	err = json.Unmarshal(body, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func getterReqBody(r *http.Request) (string, string) {
	var mName, mValue string
	uri := r.RequestURI
	mName = strings.Split(uri, "/")[3]
	mValue = strings.Split(uri, "/")[4]
	return mName, mValue
}

func notUsed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Uknown type", http.StatusNotImplemented)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Found", http.StatusNotFound)
}

func runServer() {
	hndlr := chi.NewRouter()
	hndlr.Post("/update/gauge/{metricName}/{metricValue}", gaugeHndlr)
	hndlr.Post("/update/counter/{metricName}/{metricValue}", counterHndlr)
	hndlr.Post("/update/{metricName}/", notFound)
	hndlr.Post("/update/*", notUsed)
	hndlr.Get("/value/*", getterMetricHandler)
	hndlr.Get("/", getterAllHndlr)
	//inc 4: next step
	hndlr.Post("/update/", jsonUpdMetricsHandler)
	hndlr.Post("/value/", jsonGetterMetricHandler)
	addr := fmt.Sprintf("%s:%s", host, port)
	srv := &http.Server{
		Addr:    addr,
		Handler: hndlr,
	}
	srv.SetKeepAlivesEnabled(false)
	log.Printf("waiting on port %s", port)
	log.Fatal(srv.ListenAndServe())
}

func endServer() {
	log.Println("SIGINT!")
	os.Exit(1)
}
