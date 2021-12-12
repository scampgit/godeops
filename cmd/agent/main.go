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
	reportInterval = 4
	pollInterval   = 2
)

type metric struct {
	name     string
	typename string
	value    float64
}

func (m metric) setValueInt() int64 {
	return int64(m.value)
}

func (m metric) setValueFloat() float64 {
	return m.value
}

func metricGauge(name string, value float64) metric {
	return metric{name: name, typename: "gauge", value: value}
}

func metricCounter(name string, value float64) metric {
	return metric{name: name, typename: "counter", value: value}
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

	client := http.Client{Timeout: pollInterval}
	ticker := time.NewTicker(pollInterval)
	PollCount = 0
	for {
		//log.Println("for started")
		select {
		case <-ticker.C:
			runtime.ReadMemStats(&rtm)
			PollCount++
			//fmt.Printf("in ticker poll: %d & reportint: %d", PollCount, reportInterval)
			metrics := []metric{
				metricGauge("Frees", float64(rtm.Alloc)),
				metricGauge("Frees", float64(rtm.Alloc)),
				metricGauge("Frees", float64(rtm.Alloc)),
				metricGauge("Frees", float64(rtm.Alloc)),
				metricGauge("BuckHashSys", float64(rtm.BuckHashSys)),
				metricGauge("Frees", float64(rtm.Frees)),
				metricGauge("GCCPUFraction", float64(rtm.GCCPUFraction)),
				metricGauge("GCSys", float64(rtm.GCSys)),
				metricGauge("HeapAlloc", float64(rtm.HeapAlloc)),
				metricGauge("HeapIdle", float64(rtm.HeapIdle)),
				metricGauge("HeapInuse", float64(rtm.HeapInuse)),
				metricGauge("HeapObjects", float64(rtm.HeapObjects)),
				metricGauge("HeapReleased", float64(rtm.HeapReleased)),
				metricGauge("HeapSys", float64(rtm.HeapSys)),
				metricGauge("LastGC", float64(rtm.LastGC)),
				metricGauge("Lookups", float64(rtm.Lookups)),
				metricGauge("MCacheInuse", float64(rtm.MCacheInuse)),
				metricGauge("MCacheSys", float64(rtm.MCacheSys)),
				metricGauge("MSpanInuse", float64(rtm.MSpanInuse)),
				metricGauge("MSpanSys", float64(rtm.MSpanSys)),
				metricGauge("Mallocs", float64(rtm.Mallocs)),
				metricGauge("NextGC", float64(rtm.NextGC)),
				metricGauge("NumForcedGC", float64(rtm.NumForcedGC)),
				metricGauge("NumGC", float64(rtm.NumGC)),
				metricGauge("OtherSys", float64(rtm.OtherSys)),
				metricGauge("PauseTotalNs", float64(rtm.PauseTotalNs)),
				metricGauge("StackInuse", float64(rtm.StackInuse)),
				metricGauge("StackSys", float64(rtm.StackSys)),
				metricGauge("Sys", float64(rtm.Sys)),
				metricGauge("RandomValue", rand.Float64()*1000),
			}
			metricCounter := metricCounter("PollCount", float64(PollCount))

			if PollCount == reportInterval {
				for _, i := range metrics {
					i.senderDatas(&client)
				}
				metricCounter.senderDatas(&client)
				client.CloseIdleConnections()
				PollCount = 0
			}
		case <-ctx.Done():
			fmt.Println("Context is canselled.")
			return
		}
	}
}

func (m metric) senderDatas(client *http.Client) error {
	const (
		gType     = "gauge"
		countType = "counter"
	)
	var url string
	if m.typename == gType {
		url = fmt.Sprintf("http://%s:%s/update/%s/%s/%f", host, port, m.typename, m.name, m.setValueFloat())
	} else if m.typename == countType {
		url = fmt.Sprintf("http://%s:%s/update/%s/%s/%d", host, port, m.typename, m.name, m.setValueInt())
	}
	resp, err := client.Post(url, "text/plain", nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()
	return nil
}
