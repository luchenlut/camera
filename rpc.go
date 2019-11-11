package camera

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"iot-hub/api"
	"sync"
	"time"
)

var asset api.InternalClient

var command api.ManagerClient

var once sync.Once

func NewAssetRPCClient(server string) {
	once.Do(func() {
		var err error
		var conn *grpc.ClientConn
		ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)

		conn, err = grpc.DialContext(ctx, server, grpc.WithInsecure())

		if err != nil || conn == nil {
			logrus.Fatalf("connect internal client fail %s,%v", server, err)
		}
		asset = api.NewInternalClient(conn)
		command = api.NewManagerClient(conn)
	})
}
