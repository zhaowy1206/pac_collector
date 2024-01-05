package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Define a global variable for the time interval.
var interval = 10 * time.Second

var (
	meter       = otel.Meter("systemUsage")
	cpuUsage    metric.Int64ObservableGauge
	memoryUsage metric.Int64ObservableGauge
)

func init() {
	if _, err := meter.Int64ObservableGauge("cpuUsage",
		metric.WithDescription(
			"The CPU usage",
		),
		metric.WithUnit("percentage"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			percent, err := cpu.Percent(0, false)
			if err != nil {
				log.Printf("Failed to get CPU usage: %v", err)
				return nil
			}
			o.Observe(int64(percent[0]))
			return nil
		}),
	); err != nil {
		panic(err)
	}

	if _, err := meter.Int64ObservableGauge("memoryUsage",
		metric.WithDescription(
			"The memory usage",
		),
		metric.WithUnit("percentage"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			v, err := mem.VirtualMemory()
			if err != nil {
				log.Printf("Failed to get memory usage: %v", err)
				return nil
			}
			o.Observe(int64(v.UsedPercent))
			return nil
		}),
	); err != nil {
		panic(err)
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	serviceName := "PAC Metrics"
	serviceVersion := "1.0"
	otelShutdown, err := setupOTelSDK(ctx, serviceName, serviceVersion)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// Run recordCPUUsage in a goroutine.
	go serveMetrics()

	// Wait for interruption.
	<-ctx.Done()
	// Stop receiving signal notifications as soon as possible.
	stop()

	return
}
