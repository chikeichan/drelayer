package protocol

import (
	"context"
	apiv1 "ddrp-relayer/protocol/v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func DialRPC(addr string) (apiv1.DDRPv1Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to DDRP instance")
	}
	client := apiv1.NewDDRPv1Client(conn)
	if _, err := client.GetStatus(context.Background(), &apiv1.Empty{}); err != nil {
		return nil, errors.Wrap(err, "healthcheck with DDRP instance failed")
	}
	return client, nil
}
