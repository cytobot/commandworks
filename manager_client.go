package commandworks

import (
	"context"

	pb "github.com/cytobot/rpc/manager"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type managerClient struct {
	client pb.ManagerClient
}

func newManagerClient(managerAddress string) (*managerClient, error) {
	conn, err := grpc.Dial(managerAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &managerClient{
		client: pb.NewManagerClient(conn),
	}, nil
}

func (c *managerClient) getWorkerStatus() ([]*pb.HealthCheckStatus, error) {
	statusList, err := c.client.GetWorkerHealthChecks(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return statusList.GetHealthChecks(), nil
}

func (c *managerClient) getListenerStatus() ([]*pb.HealthCheckStatus, error) {
	statusList, err := c.client.GetListenerHealthChecks(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return statusList.GetHealthChecks(), nil
}
