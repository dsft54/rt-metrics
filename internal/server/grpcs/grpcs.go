// Package grpcs defines grpc server functions for metrics server
package grpcs

import (
	"context"
	"fmt"

	"github.com/dsft54/rt-metrics/internal/server/storage"
	pb "github.com/dsft54/rt-metrics/proto"
)

type MetricsServer struct {
	pb.UnimplementedMetricServer

	Storage storage.IStorage
}

func (ms *MetricsServer) AddMetric(ctx context.Context, in *pb.AddMetricRequest) (*pb.AddMetricResponse, error) {
	var response pb.AddMetricResponse
	inM := &storage.Metrics{
		ID:    in.Metrics.Id,
		MType: in.Metrics.Mtype,
		Delta: &in.Metrics.Delta,
		Value: &in.Metrics.Value,
		Hash:  in.Metrics.Hash,
	}
	err := ms.Storage.InsertMetric(inM)
	if err != nil {
		response.Error = fmt.Sprint("failed to add metric: ", inM)
		return &response, err
	}
	return &response, nil
}

func (ms *MetricsServer) AddMetrics(ctx context.Context, in *pb.AddMetricsRequest) (*pb.AddMetricsResponse, error) {
	var response pb.AddMetricsResponse
	ainM := []storage.Metrics{}
	for _, v := range in.Metrics {
		inM := &storage.Metrics{
			ID:    v.Id,
			MType: v.Mtype,
			Delta: &v.Delta,
			Value: &v.Value,
			Hash:  v.Hash,
		}
		ainM = append(ainM, *inM)
	}
	err := ms.Storage.InsertBatchMetric(ainM)
	if err != nil {
		response.Error = fmt.Sprint("failed to add metric: ", ainM)
		return &response, err
	}
	return &response, nil
}
