package sender

import (
	"github.com/gam6itko/go-musthave-metrics/internal/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

type GRPCSender struct {
	conn *grpc.ClientConn
}

func NewGRPCSender(address string) *GRPCSender {
	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}

	return &GRPCSender{
		conn: conn,
	}
}

func (ths *GRPCSender) Send([]*common.Metrics) error {

	return nil
}
