package main

import (
	"fmt"
	"os"

	"github.com/hyperhq/kubestack/pkg/types"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	server = "10.10.101.79:4237"
)

func main() {
	conn, err := grpc.Dial(server, grpc.WithInsecure())

	if err != nil {
		fmt.Printf("Connect server error: %v", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("Connect succes!")
	client := types.NewFloatingIPsClient(conn)

	request := types.DelFloatingIpRequest{
		FloatingipId: "b2391f12-42da-4cc0-bb69-8fd55428236c",
	}
	response, err := client.DelFloatingIp(context.Background(), &request)
	if err != nil {
		fmt.Printf(" Del error :%v", err)
		os.Exit(1)
	}
	fmt.Printf("Delete FloatingIp: %v", response)
	//	client := types.NewNetworksClient(conn)
	//	request := types.CheckTenantIDRequest{
	//		TenantID: "03b9174c12664918ac4323c04bf60125",
	//	}

	//	response, err := client.CheckTenantID(context.Background(), &request)
	//	if err != nil {
	//		fmt.Printf("CheckTenantId error: %v", err)
	//		os.Exit(1)
	//	}

	//	fmt.Printf("Got response : %v", response)

	//	request := types.GetNetworkRequest{
	//		Name: "testnet",
	//	}

	//	response, err := client.GetNetwork(context.Background(), &request)
	//	if err != nil {
	//		fmt.Printf("Get Network error: %v", err)
	//		os.Exit(1)
	//	}

	//	fmt.Printf("Get network response: %v", response)

	//	client := types.NewPublicAPIClient(conn)
	//	request := types.PodInfoRequest{
	//		PodID: "pod-zpIOTSAjmM",
	//	}
	//	response, err := client.PodInfo(context.Background(), &request)
	//	if err != nil {
	//		fmt.Printf("Get PodInfo error: %v", err)
	//		os.Exit(1)
	//	}

	//	fmt.Printf("Got response: %v", response)
}
