package sender

import "github.com/gam6itko/go-musthave-metrics/internal/common"

type ISender interface {
	Send([]*common.Metrics) error
}
