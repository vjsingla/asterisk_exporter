package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

var activeChannels, activeCalls, callProcessed prometheus.Gauge

func recordMetrics() {
	go func() {
		for {
			aChannels, err := executeCommand("asterisk -rx 'core show channels' | grep 'active channels' | awk '{print $1}'")
			if err != nil {
				fmt.Println(err)
			}
			activeChannels.Set(aChannels)

			aCalls, err := executeCommand("asterisk -rx 'core show channels' | grep 'active calls' | awk '{print $1}'")
			if err != nil {
				fmt.Println(err)
			}
			activeCalls.Set(aCalls)

			cProcessed, err := executeCommand("asterisk -rx 'core show channels' | grep 'calls processed' | awk '{print $1}'")
			if err != nil {
				fmt.Println(err)
			}
			callProcessed.Set(cProcessed)
		}
	}()
}

func executeCommand(cmdStr string) (float64, error) {
	var o float64
	cmd := exec.Command("sh", "-c", cmdStr)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return o, err
	}

	o, err = strconv.ParseFloat(string(stdoutStderr), 8)
	if err != nil {
		return o, err
	}

	return o, nil

}

func main() {
	h, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	activeChannels = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "asterisk_active_channels",
		Help: "Total current active channels",
		ConstLabels: map[string]string{
			"type": "active channels",
			"host": h,
		},
	})
	activeCalls = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "asterisk_active_calls",
		Help: "Total current active calls",
		ConstLabels: map[string]string{
			"type": "active calls",
			"host": h,
		},
	})
	callProcessed = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "asterisk_calls_processed",
		Help: "Total current calls processed",
		ConstLabels: map[string]string{
			"type": "calls processed",
			"host": h,
		},
	})

	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9200", nil)
}
