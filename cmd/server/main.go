package main

import (
	"fmt"
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
	host = "127.0.0.1"
	port = "8080"
)

type gauge struct {
	v float64
}

type counter struct {
	v int64
}

var gmetrics = make(map[string]gauge)
var cmetrics = make(map[string]counter)
var templateDataMap = make(map[string]interface{})

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

func gaugeHndlr(w http.ResponseWriter, r *http.Request) {
	mName, mValue := getterReqBody(r)
	value, err := strconv.ParseFloat(string(mValue), 64)
	if err != nil {
		http.Error(w, "parsing error. Bad request", http.StatusBadRequest)
	} else {
		gmetrics[mName] = gauge{v: value}
		w.WriteHeader(http.StatusOK)
	}
}

func counterHndlr(w http.ResponseWriter, r *http.Request) {
	mName, mVal := getterReqBody(r)
	value, err := strconv.ParseInt(string(mVal), 10, 64)
	if err != nil {
		http.Error(w, "parsing error", http.StatusBadRequest)
		return
	}
	cmetrics[mName] = counter{cmetrics[mName].v + value}
	w.WriteHeader(http.StatusOK)
}

func getterMetricHndlr(w http.ResponseWriter, r *http.Request) {
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
	htmlPage, err := os.ReadFile("metrics.html")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	tmpl := template.Must(template.New("").Parse(string(htmlPage)))
	tmpl.Execute(w, templateDataMap)
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
	hndlr.Get("/value/*", getterMetricHndlr)
	hndlr.Get("/", getterAllHndlr)
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
