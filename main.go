package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getTerminalSecurityDuration(terminal int) (float64, error) {
	curl := fmt.Sprintf("curl -s https://www.dublinairport.com/flight-information/live-departures | grep \"Terminal</span>\\ T%v\" | awk {'print $5'}", terminal)
	out, err := exec.Command("bash", "-c", curl).Output()
	if err != nil {
		return 0, err
	}

	outSanitised := strings.TrimSuffix(string(out), "\n")

	f, err := strconv.ParseFloat(outSanitised, 64)
	if err != nil {
		return 0, err
	}

	return f, nil

}

func recordMetrics() {
	go func() {
		for {

			t1Duration, err := getTerminalSecurityDuration(1)
			if err != nil {
				fmt.Printf("Error getting T1 duration: %s", err)
			}
			t2Duration, err := getTerminalSecurityDuration(2)
			if err != nil {
				fmt.Printf("Error getting T2 duration: %s", err)
			}

			securityTime.With(
				prometheus.Labels{
					"terminal": "T1",
				},
			).Set(t1Duration)

			securityTime.With(
				prometheus.Labels{
					"terminal": "T2",
				},
			).Set(t2Duration)

			time.Sleep(1 * time.Minute)
		}
	}()
}

var (
	securityTime = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dubair_security_queue_time",
			Help: "How long security time takes in Dublin airport",
		},
		[]string{"terminal"},
	)
)

func main() {
	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
