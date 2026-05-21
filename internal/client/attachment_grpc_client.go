package client

import (
	"ne_noy/proto/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewAttachmentGRPCClient(addr string) (gen.AttachmentServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return gen.NewAttachmentServiceClient(conn), conn, nil
}
