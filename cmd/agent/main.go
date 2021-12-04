package main

import (
	"context"
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
	host           = "127.0.0.1"
	port           = "8080"
	ReportInterval = 4
)

type Metric struct {
	name     string
	typename string
	value    float64
}

func (m Metric) setValueInt() int64 {
	return int64(m.value)
}

func (m Metric) setValueFloat() float64 {
	return m.value
}

func MetricGauge(name string, value float64) Metric {
	return Metric{name: name, typename: "gauge", value: value}
}

func MetricCounter(name string, value float64) Metric {
	return Metric{name: name, typename: "counter", value: value}
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
	var rtm runtime.MemStats
	var PollCount int64

	client := http.Client{Timeout: 2 * time.Second}
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			runtime.ReadMemStats(&rtm)
			PollCount += 1
			metrics := []Metric{
				MetricGauge("Frees", float64(rtm.Alloc)),
				MetricGauge("Frees", float64(rtm.Alloc)),
				MetricGauge("BuckHashSys", float64(rtm.BuckHashSys)),
				MetricGauge("Frees", float64(rtm.Frees)),
				MetricGauge("GCCPUFraction", float64(rtm.GCCPUFraction)),
				MetricGauge("GCSys", float64(rtm.GCSys)),
				MetricGauge("HeapAlloc", float64(rtm.HeapAlloc)),
				MetricGauge("HeapIdle", float64(rtm.HeapIdle)),
				MetricGauge("HeapInuse", float64(rtm.HeapInuse)),
				MetricGauge("HeapObjects", float64(rtm.HeapObjects)),
				MetricGauge("HeapReleased", float64(rtm.HeapReleased)),
				MetricGauge("HeapSys", float64(rtm.HeapSys)),
				MetricGauge("LastGC", float64(rtm.LastGC)),
				MetricGauge("Lookups", float64(rtm.Lookups)),
				MetricGauge("MCacheInuse", float64(rtm.MCacheInuse)),
				MetricGauge("MCacheSys", float64(rtm.MCacheSys)),
				MetricGauge("MSpanInuse", float64(rtm.MSpanInuse)),
				MetricGauge("MSpanSys", float64(rtm.MSpanSys)),
				MetricGauge("Mallocs", float64(rtm.Mallocs)),
				MetricGauge("NextGC", float64(rtm.NextGC)),
				MetricGauge("NumForcedGC", float64(rtm.NumForcedGC)),
				MetricGauge("NumGC", float64(rtm.NumGC)),
				MetricGauge("OtherSys", float64(rtm.OtherSys)),
				MetricGauge("PauseTotalNs", float64(rtm.PauseTotalNs)),
				MetricGauge("StackInuse", float64(rtm.StackInuse)),
				MetricGauge("StackSys", float64(rtm.StackSys)),
				MetricGauge("Sys", float64(rtm.Sys)),
				MetricGauge("RandomValue", rand.Float64()*1000),
			}
			MetricCounter := MetricCounter("PollCount", float64(PollCount))

			if PollCount == ReportInterval {
				for _, i := range metrics {
					i.senderDatas(&client)
				}
				MetricCounter.senderDatas(&client)
				client.CloseIdleConnections()
				PollCount = 0
			}
		case <-ctx.Done():
			fmt.Println("Context is canselled.")
			return
		}
	}
}

func (m Metric) senderDatas(client *http.Client) (bool, error) {
	var url string
	if m.typename == "gauge" {
		url = fmt.Sprintf("http://%s:%s/update/%s/%s/%f", host, port, m.typename, m.name, m.setValueFloat())
	} else if m.typename == "counter" {
		url = fmt.Sprintf("http://%s:%s/update/%s/%s/%d", host, port, m.typename, m.name, m.setValueInt())
	}
	resp, err := client.Post(url, "text/plain", nil)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer resp.Body.Close()
	return true, nil
}
