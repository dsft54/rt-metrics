// Package grpcc provides client functions for metrics server

package grpcc

import (
	"context"
	"log"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/dsft54/rt-metrics/internal/agent/storage"
	"github.com/dsft54/rt-metrics/internal/server/grpcs"
	pb "github.com/dsft54/rt-metrics/proto"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterMetricServer(s, &grpcs.MetricsServer{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestSendMetric(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := pb.NewMetricClient(conn)
	connE, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer connE.Close()
	cE := pb.NewMetricClient(connE)
	var (
		d int64 = 3
		v       = 3.14
	)
	type args struct {
		ctx context.Context
		c   pb.MetricClient
		m   storage.Metrics
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal g",
			args: args{
				ctx: context.Background(),
				c:   c,
				m: storage.Metrics{
					MType: "gauge",
					ID:    "Alloc",
					Value: &v,
				},
			},
			wantErr: false,
		},
		{
			name: "normal c",
			args: args{
				ctx: context.Background(),
				c:   c,
				m: storage.Metrics{
					MType: "counter",
					ID:    "Alloc",
					Delta: &d,
				},
			},
			wantErr: false,
		},
		{
			name: "err c",
			args: args{
				ctx: context.Background(),
				c:   cE,
				m: storage.Metrics{
					MType: "counter",
					ID:    "Alloc",
					Delta: &d,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendMetric(tt.args.ctx, tt.args.c, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("SendMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendMetrics(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := pb.NewMetricClient(conn)
	connE, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer connE.Close()
	cE := pb.NewMetricClient(connE)
	var (
		d int64 = 3
		v       = 3.14
	)
	type args struct {
		ctx context.Context
		c   pb.MetricClient
		mb  []storage.Metrics
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "err g",
			args: args{
				ctx: context.Background(),
				c:   cE,
				mb: []storage.Metrics{
					{
						MType: "gauge",
						ID:    "Alloc",
						Value: &v,
					},
					{
						MType: "counter",
						ID:    "Alloc",
						Delta: &d,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "err g",
			args: args{
				ctx: context.Background(),
				c:   c,
				mb: []storage.Metrics{
					{
						MType: "gauge",
						ID:    "Alloc",
						Value: &v,
					},
					{
						MType: "counter",
						ID:    "Alloc",
						Delta: &d,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendMetrics(tt.args.ctx, tt.args.c, tt.args.mb); (err != nil) != tt.wantErr {
				t.Errorf("SendMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
