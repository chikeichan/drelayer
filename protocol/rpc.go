package protocol

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func DialRPC(addr string) (DDRPClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to DDRP instance")
	}
	client := NewDDRPClient(conn)
	if _, err := client.GetStatus(context.Background(), &empty.Empty{}); err != nil {
		return nil, errors.Wrap(err, "healthcheck with DDRP instance failed")
	}
	return client, nil
}
