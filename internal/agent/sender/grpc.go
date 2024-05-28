package sender

import (
	"context"
	"errors"
	"github.com/gam6itko/go-musthave-metrics/internal/common"
	"github.com/gam6itko/go-musthave-metrics/proto"
	"google.golang.org/grpc"
)

type GRPCSender struct {
	client proto.MetricsClient
}

func NewGRPCSender(conn *grpc.ClientConn) *GRPCSender {
	client := proto.NewMetricsClient(conn)

	return &GRPCSender{
		client: client,
	}
}

func (ths *GRPCSender) Send(ctx context.Context, metrics []*common.Metrics) error {

	for _, m := range metrics {
		switch m.MType {
		case "counter":
			req := proto.CounterIncRequest{
				Name:  m.ID,
				Value: *m.Delta,
			}
			resp, err := ths.client.CounterInc(ctx, &req)
			if err != nil {
				return err
			}
			if resp.Error != "" {
				return errors.New(resp.Error)
			}
		case "gauge":
			req := proto.GaugeSetRequest{
				Name:  m.ID,
				Value: *m.Value,
			}
			resp, err := ths.client.GaugeSet(ctx, &req)
			if err != nil {
				return err
			}
			if resp.Error != "" {
				return errors.New(resp.Error)
			}
		}
	}

	return nil
}
