package sender

import (
	"context"
	"github.com/gam6itko/go-musthave-metrics/internal/common"
)

type ISender interface {
	Send(context.Context, []*common.Metrics) error
}
