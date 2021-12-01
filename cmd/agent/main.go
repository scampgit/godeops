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
	host = "127.0.0.1"
	port = "8080"
)

func main() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())
	go getterMetrics(ctx)
	<-signalChannel
	cancel()
	endAgent()
}

func endAgent() {
	log.Println("SIGINT!")
	os.Exit(1)
}

func getterMetrics(ctx context.Context) {
	var PollCount int
	var rtmemmetrics runtime.MemStats

	client := http.Client{Timeout: 2 * time.Second}
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			runtime.ReadMemStats(&rtmemmetrics)
			PollCount += 1
			metrics := []Metric{
				newMetricGauge("RandomValue", rand.Float64()*100),
			}
			metricCounter := newMetricCounter("PollCount", float64(PollCount))

		case <-ctx.Done():
			fmt.Println("Ctx has canceled successfully.")
			return
		}
	}
}
