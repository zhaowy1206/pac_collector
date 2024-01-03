package main

import (
	"context"
	"log"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

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

func recordUsage(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			continue
		}
	}
}
