package main

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

type multiplyByTwoProcessor struct {
	nextConsumer consumer.Metrics
}

func newMultiplyByTwoProcessor(nextConsumer consumer.Metrics) *multiplyByTwoProcessor {
	return &multiplyByTwoProcessor{nextConsumer: nextConsumer}
}

func (p *multiplyByTwoProcessor) ConsumeMetrics(ctx context.Context, md pdata.Metrics) error {
	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)
		ilms := rm.InstrumentationLibraryMetrics()
		for j := 0; j < ilms.Len(); j++ {
			ilm := ilms.At(j)
			metrics := ilm.Metrics()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)
				dataPoints := metric.DoubleSum().DataPoints()
				for l := 0; l < dataPoints.Len(); l++ {
					dataPoint := dataPoints.At(l)
					dataPoint.SetValue(dataPoint.Value() * 2)
				}
			}
		}
	}
	return p.nextConsumer.ConsumeMetrics(ctx, md)
}

func (p *multiplyByTwoProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func main() {
	factory := processorhelper.NewFactory(
		"multiply_by_two",
		func() component.Processor {
			return newMultiplyByTwoProcessor
		},
	)
	processorhelper.StartProcessor(factory)
}
