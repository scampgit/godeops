package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const (
	host                = "127.0.0.1"
	port                = "8080"
	reportInterval      = 4
	pollInterval        = 2
	metric_type_gauge   = "gauge"
	metric_type_counter = "counter"
)

//type metric struct {
//	name     string
//	typename string
//	value    float64
//}

//func (m metric) setValueInt() int64 {
//	return int64(m.value)
//}

//func (m metric) setValueFloat() float64 {
//	return m.value
//}

//func metricGauge(name string, value float64) metric {
//	return metric{name: name, typename: "gauge", value: value}
//}

//func metricCounter(name string, value float64) metric {
//	return metric{name: name, typename: "counter", value: value}
//}

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

var metrics = make(map[string]Metric)

func jMetricGauge(name string, value float64) {
	metrics[name] = Metric{ID: name, MType: metric_type_gauge, Value: &value}
}

func jMetricCounter(name string, value int64) {
	metrics[name] = Metric{ID: name, MType: metric_type_counter, Delta: &value}
}

func main() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())
	go getterDatas(ctx)
	<-signalChannel
	cancel()
	endAgent()
}

func endAgent() {
	log.Println("SIGINT!")
	os.Exit(1)
}

func getterDatas(ctx context.Context) {
	const (
		pollInterval = time.Second * 2
	)
	var rtm runtime.MemStats
	var PollCount int64

	ticker := time.NewTicker(pollInterval)
	PollCount = 0
	for {
		//log.Println("for started")
		select {
		case <-ticker.C:
			runtime.ReadMemStats(&rtm)
			PollCount++
			//fmt.Printf("in ticker poll: %d & reportint: %d", PollCount, reportInterval)
			jMetricGauge("Alloc", float64(rtm.Alloc))
			jMetricGauge("Alloc", float64(rtm.Alloc))
			jMetricGauge("TotalAlloc", float64(rtm.TotalAlloc))
			jMetricGauge("BuckHashSys", float64(rtm.BuckHashSys))
			jMetricGauge("Frees", float64(rtm.Frees))
			jMetricGauge("GCCPUFraction", float64(rtm.GCCPUFraction))
			jMetricGauge("GCSys", float64(rtm.GCSys))
			jMetricGauge("HeapAlloc", float64(rtm.HeapAlloc))
			jMetricGauge("HeapIdle", float64(rtm.HeapIdle))
			jMetricGauge("HeapInuse", float64(rtm.HeapInuse))
			jMetricGauge("HeapObjects", float64(rtm.HeapObjects))
			jMetricGauge("HeapReleased", float64(rtm.HeapReleased))
			jMetricGauge("HeapSys", float64(rtm.HeapSys))
			jMetricGauge("LastGC", float64(rtm.LastGC))
			jMetricGauge("Lookups", float64(rtm.Lookups))
			jMetricGauge("MCacheInuse", float64(rtm.MCacheInuse))
			jMetricGauge("MCacheSys", float64(rtm.MCacheSys))
			jMetricGauge("MSpanInuse", float64(rtm.MSpanInuse))
			jMetricGauge("MSpanSys", float64(rtm.MSpanSys))
			jMetricGauge("Mallocs", float64(rtm.Mallocs))
			jMetricGauge("NextGC", float64(rtm.NextGC))
			jMetricGauge("NumForcedGC", float64(rtm.NumForcedGC))
			jMetricGauge("NumGC", float64(rtm.NumGC))
			jMetricGauge("OtherSys", float64(rtm.OtherSys))
			jMetricGauge("PauseTotalNs", float64(rtm.PauseTotalNs))
			jMetricGauge("StackInuse", float64(rtm.StackInuse))
			jMetricGauge("StackSys", float64(rtm.StackSys))
			jMetricGauge("Sys", float64(rtm.Sys))
			jMetricGauge("RandomValue", rand.Float64()*100)
			jMetricCounter("PollCount", int64(PollCount))
			if PollCount == reportInterval {
				metricsSnapshot := metrics
				for _, m := range metricsSnapshot {
					err := updateMetric(&m)
					if err != nil {
						log.Println(err)
					}
					if m.ID == "PollCount" {
						PollCount = 0
						m.Delta = &PollCount
					}
				}
			}
		case <-ctx.Done():
			fmt.Println("Context is canselled.")
			return
		}
	}
}

func updateMetric(m *Metric) error {
	var url string

	mSer, err := json.Marshal(*m)
	if err != nil {
		return err
	}
	url = fmt.Sprintf("http://%s:%s/update/", host, port)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(mSer))

	if err != nil {
		log.Println(err)
		return err
	}

	defer resp.Body.Close()
	return err
}
