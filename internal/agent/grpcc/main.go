// Package grpcc provides client functions for metrics server
package grpcc

import (
	"context"
	"log"

	"github.com/dsft54/rt-metrics/internal/agent/storage"
	pb "github.com/dsft54/rt-metrics/proto"
)

func SendMetric(ctx context.Context, c pb.MetricClient, m storage.Metrics) error {
	pbMetric := pb.Metrics{
		Id:    m.ID,
		Mtype: m.MType,
		Hash:  m.Hash,
	}
	if m.Delta != nil {
		pbMetric.Delta = *m.Delta
	}
	if m.Value != nil {
		pbMetric.Value = *m.Value
	}
	resp, err := c.AddMetric(ctx, &pb.AddMetricRequest{Metrics: &pbMetric})
	if err != nil {
		return err
	}
	log.Println(resp)
	return nil
}

func SendMetrics(ctx context.Context, c pb.MetricClient, mb []storage.Metrics) error {
	pbMetrics := []*pb.Metrics{}
	for _, m := range mb {
		pbMetric := pb.Metrics{
			Id:    m.ID,
			Mtype: m.MType,
			Hash: m.Hash,
		}
		if m.Delta != nil {
			pbMetric.Delta = *m.Delta
		}
		if m.Value != nil {
			pbMetric.Value = *m.Value
		}
		pbMetrics = append(pbMetrics, &pbMetric)
	}
	resp, err := c.AddMetrics(ctx, &pb.AddMetricsRequest{Metrics: pbMetrics})
	if err != nil {
		return err
	}
	log.Println(resp)
	return nil
}
