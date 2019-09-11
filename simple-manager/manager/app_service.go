package manager

import (
	"context"
	"fmt"
	"log"
	"net"

	rpc "github.com/r2-studio/robotmon-desktop/simple-manager/manager/apprpc"
	"google.golang.org/grpc"
)

type AppService struct {
	adbClient *AdbClient
}

func NewAppService(adbClient *AdbClient) *AppService {
	a := &AppService{
		adbClient: adbClient,
	}
	a.Init()
	return a
}

func (a *AppService) Init() {
	opt1 := grpc.MaxRecvMsgSize(64 * 1024 * 1024)
	opt2 := grpc.MaxSendMsgSize(64 * 1024 * 1024)
	grpcServer := grpc.NewServer(opt1, opt2)

	rpc.RegisterAppServiceServer(grpcServer, a)
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Println("Listen :9487 for gRPC")
	go grpcServer.Serve(lis)
}

func (a *AppService) GetDevices(context.Context, *rpc.Empty) (*rpc.Devices, error) {
	serials, err := adbClient.GetDevices()
	if err != nil {
		return nil, err
	}
	devices := make([]*rpc.Device, 0, len(serials))
	for _, serial := range serials {
		device := &rpc.Device{
			Serial: serial,
		}
		devices = append(devices, device)
	}
	result := &rpc.Devices{
		Devices: devices,
	}
	return result, nil
}

func (a *AppService) AdbConnect(context.Context, *rpc.AdbConnectParams) (*rpc.Message, error) {
	return nil, nil
}

func (a *AppService) AdbShell(context.Context, *rpc.AdbShellParams) (*rpc.Message, error) {
	return nil, nil
}
