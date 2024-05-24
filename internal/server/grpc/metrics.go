package grpc

import (
	"context"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage"
	"github.com/gam6itko/go-musthave-metrics/proto"
)

type MetricsServerImpl struct {
	proto.UnimplementedMetricsServer

	storage storage.IStorage
}

func NewMetricsServerImpl(storage storage.IStorage) *MetricsServerImpl {
	return &MetricsServerImpl{storage: storage}
}

func (ths MetricsServerImpl) CounterInc(ctx context.Context, req *proto.CounterIncRequest) (*proto.CounterIncResponse, error) {
	resp := &proto.CounterIncResponse{}
	err := ths.storage.CounterInc(ctx, req.Name, req.Value)
	if err != nil {
		resp.Error = err.Error()
	}
	return resp, nil
}

func (ths MetricsServerImpl) CounterGet(ctx context.Context, req *proto.CounterGetRequest) (*proto.CounterGetResponse, error) {
	resp := &proto.CounterGetResponse{}
	val, err := ths.storage.CounterGet(ctx, req.Name)
	if err != nil {
		resp.Error = err.Error()
	}

	resp.Value = val
	return resp, nil
}

func (ths MetricsServerImpl) GaugeSet(ctx context.Context, req *proto.GaugeSetRequest) (*proto.GaugeSetResponse, error) {
	resp := &proto.GaugeSetResponse{}
	err := ths.storage.GaugeSet(ctx, req.Name, req.Value)
	if err != nil {
		resp.Error = err.Error()
	}
	return resp, nil
}

func (ths MetricsServerImpl) GaugeGet(ctx context.Context, req *proto.GaugeGetRequest) (*proto.GaugeGetResponse, error) {
	resp := &proto.GaugeGetResponse{}
	val, err := ths.storage.GaugeGet(ctx, req.Name)
	if err != nil {
		resp.Error = err.Error()
	}

	resp.Value = val
	return resp, nil
}
