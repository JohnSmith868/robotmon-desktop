package manager

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	rpc "github.com/r2-studio/robotmon-desktop/simple-manager/manager/apprpc"
	"github.com/r2-studio/robotmon-desktop/simple-manager/manager/proxy"
	"google.golang.org/grpc"
)

// AppService app service struct
type AppService struct {
	adbClient *AdbClient
	// grpcAddress -> httpBindingAddress
	proxyTable map[string]string
}

// NewAppService new app service
func NewAppService(adbClient *AdbClient) *AppService {
	a := &AppService{
		adbClient:  adbClient,
		proxyTable: make(map[string]string),
	}
	a.Init()
	return a
}

// Init init app service
func (a *AppService) Init() {
	opt1 := grpc.MaxRecvMsgSize(64 * 1024 * 1024)
	opt2 := grpc.MaxSendMsgSize(64 * 1024 * 1024)
	grpcServer := grpc.NewServer(opt1, opt2)

	rpc.RegisterAppServiceServer(grpcServer, a)
	lis, err := net.Listen("tcp", ":9487")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	fmt.Println("Listen :9487 for gRPC")
	go grpcServer.Serve(lis)

	time.Sleep(time.Second)

	fmt.Println("Listen :9488 for http proxy")
	go proxy.RunProxy("0.0.0.0:9488", "localhost:9487")
}

// GetDevices get devices
func (a *AppService) GetDevices(context.Context, *rpc.Empty) (*rpc.Devices, error) {
	forwardResult := adbClient.ForwardList()
	forwards := strings.Split(forwardResult, "\n")

	serials, err := adbClient.GetDevices()
	if err != nil {
		return nil, err
	}
	devices := make([]*rpc.Device, 0, len(serials))
	for _, serial := range serials {
		// find forward port
		forward := ""
		for _, f := range forwards {
			if strings.Contains(f, serial) && strings.Contains(f, "tcp:8081") {
				forward = f
				break
			}
		}
		// find ip
		ip := adbClient.GetIPAddress(serial)
		launched := false
		pid1 := ""
		pid2 := ""
		pids, err := adbClient.GetPids(serial, "app_process")
		if err == nil {
			if len(pids) > 0 {
				pid1 = pids[0]
			}
			if len(pids) > 1 {
				pid2 = pids[1]
				launched = true
			}
		}
		device := &rpc.Device{
			Serial:          serial,
			ServiceIp:       ip,
			ServicePort:     "8081",
			ServicePid1:     pid1,
			ServicePid2:     pid2,
			ServiceForward:  forward,
			ServiceLaunched: launched,
		}
		devices = append(devices, device)
	}
	result := &rpc.Devices{
		Devices: devices,
	}
	return result, nil
}

// GetStartCommand get devices
func (a *AppService) GetStartCommand(ctx context.Context, req *rpc.DeviceSerial) (*rpc.GetStartCommandResult, error) {
	_, details, err := adbClient.GetRobotmonStartCommand(req.Serial)
	if err != nil {
		return nil, err
	}
	if len(details) != 5 {
		return nil, fmt.Errorf("Get start command failed: %v", details)
	}
	return &rpc.GetStartCommandResult{
		LdPath:      details[0],
		ClassPath:   details[1],
		AppProcess:  details[2],
		BaseCommand: details[3],
		FullCommand: details[4],
	}, nil
}

// AdbRestart call adb kill-server then adb start-server
func (a *AppService) AdbRestart(context.Context, *rpc.Empty) (*rpc.Empty, error) {
	adbClient.Restart()
	return &rpc.Empty{}, nil
}

// AdbConnect call adb connect
func (a *AppService) AdbConnect(ctx context.Context, req *rpc.AdbConnectParams) (*rpc.Message, error) {
	result, err := adbClient.Connect(req.Ip, req.Port)
	if err != nil {
		return nil, err
	}
	return &rpc.Message{Message: result}, nil
}

// AdbShell call adb shell
func (a *AppService) AdbShell(ctx context.Context, req *rpc.AdbShellParams) (*rpc.Message, error) {
	result, err := adbClient.Shell(req.Serial, req.Command)
	if err != nil {
		return nil, err
	}
	return &rpc.Message{Message: result}, nil
}

// AdbForward call adb shell
func (a *AppService) AdbForward(ctx context.Context, req *rpc.AdbForwardParams) (*rpc.Message, error) {
	result, err := adbClient.Forward(req.Serial, req.DevicePort, req.PcPort)
	if err != nil {
		return nil, err
	}
	if result {
		return &rpc.Message{Message: fmt.Sprintf("Forward success %s -> %s", req.DevicePort, req.PcPort)}, nil
	}
	return &rpc.Message{Message: "Forward failed"}, nil
}

// AdbForwardList call adb shell
func (a *AppService) AdbForwardList(context.Context, *rpc.Empty) (*rpc.Message, error) {
	result := adbClient.ForwardList()
	return &rpc.Message{Message: result}, nil
}

// AdbTCPIP call adb shell
func (a *AppService) AdbTCPIP(ctx context.Context, req *rpc.AdbTCPIPParams) (*rpc.Message, error) {
	err := adbClient.TCPIP(req.Serial, req.Port)
	if err != nil {
		return nil, err
	}
	return &rpc.Message{Message: fmt.Sprintf("Forward success port: %s", req.Port)}, nil
}

// StartService get devices
func (a *AppService) StartService(ctx context.Context, req *rpc.DeviceSerial) (*rpc.StartServiceResult, error) {
	pids, err := adbClient.StartRobotmonService(req.Serial)
	if pids == nil || err != nil {
		return nil, err
	}
	var pid1, pid2 string
	if len(pids) > 0 {
		pid1 = pids[0]
	}
	if len(pids) > 1 {
		pid2 = pids[1]
	}
	return &rpc.StartServiceResult{
		Pid1: pid1,
		Pid2: pid2,
	}, nil
}

// StopService get devices
func (a *AppService) StopService(ctx context.Context, req *rpc.DeviceSerial) (*rpc.Message, error) {
	err := adbClient.StopService(req.Serial)
	if err != nil {
		return nil, err
	}
	return &rpc.Message{Message: "StopService success"}, nil
}

// CreateProxy get devices
func (a *AppService) CreateProxy(ctx context.Context, req *rpc.CreateGRPCProxy) (*rpc.Message, error) {
	httpAddress := req.HttpAddress
	grpcAddress := req.GrpcAddress
	_, ok := a.proxyTable[httpAddress]
	if ok {
		return &rpc.Message{Message: "Proxy already created"}, nil
	}
	a.proxyTable[httpAddress] = grpcAddress
	go func() {
		proxy.RunProxy(req.HttpAddress, req.GrpcAddress)
		delete(a.proxyTable, httpAddress)
	}()
	time.Sleep(time.Second)
	return &rpc.Message{Message: "Proxy server created"}, nil
}
