package main

import (
	"errors"
	"fmt"

	"github.com/golang/glog"
	"github.com/hyperhq/kubestack/pkg/types"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Interface interface {
	// Pods returns a pod interface
	Pods() Pods
	// Networks returns a network interface
	Networks() Networks
	//Subnets returns a subnets interface
	Subnets() Subnets
	// ProviderName returns the network provider ID.
	ProviderName() string
	// CheckTenantID
	CheckTenantID(tenantID string) (bool, error)
}

type Pods interface {
	// Setup pod
	SetupPod(podName, namespace, podInfraContainerID string, network *types.Network, containerRuntime, subnetID string) error
	// Teardown pod
	TeardownPod(podName, namespace, podInfraContainerID string, network *types.Network, containerRuntime string) error
	// Status of pod
	PodStatus(podName, namespace, podInfraContainerID string, network *types.Network, containerRuntime string) (string, error)
}

// Networks is an abstract, pluggable interface for network segment
type Networks interface {
	// Get network by networkName
	GetNetwork(networkName string) (*types.Network, error)
	// Get network by networkID
	GetNetworkByID(networkID string) (*types.Network, error)
	// Create network
	CreateNetwork(network *types.Network) error
	// Update network
	UpdateNetwork(network *types.Network) error
	// Delete network by networkName
	DeleteNetwork(networkName string) error
}

type Subnets interface {
	ListSubnets(networkID string) ([]*types.Subnet, error)
	CreateSubnet(subnet *types.Subnet) error
	DeleteSubnet(subnetID string, networkID string) error
	UpdateSubnet(subnet *types.Subnet) error
}

type NeutronProvider struct {
	server string

	networkClient      types.NetworksClient
	podClient          types.PodsClient
	loadbalancerClient types.LoadBalancersClient
	subnetClient       types.SubnetsClient
}

func InitNeutronProviders(remoteAddr string) (*NeutronProvider, error) {
	conn, err := grpc.Dial(remoteAddr, grpc.WithInsecure())
	if err != nil {
		glog.Errorf("Connect network provider %s failed: %v", remoteAddr, err)
		return nil, err
	}

	networkClient := types.NewNetworksClient(conn)
	podClient := types.NewPodsClient(conn)
	lbClient := types.NewLoadBalancersClient(conn)
	subnetClient := types.NewSubnetsClient(conn)
	resp, err := networkClient.Active(
		context.Background(),
		&types.ActiveRequest{},
	)
	if err != nil || !resp.Result {
		glog.Errorf("Active network provider %s failed: %v", remoteAddr, err)
		return nil, err
	}
	return &NeutronProvider{
		server:             remoteAddr,
		podClient:          podClient,
		loadbalancerClient: lbClient,
		networkClient:      networkClient,
		subnetClient:       subnetClient,
	}, nil
}

// Network interface is self
func (r *NeutronProvider) Networks() Networks {
	return r
}

// Pods interface is self
func (r *NeutronProvider) Pods() Pods {
	return r
}

// Subnets interface is self
func (r *NeutronProvider) Subnets() Subnets {
	return r
}

// Get network by networkName
func (r *NeutronProvider) GetNetwork(networkName string) (*types.Network, error) {
	if networkName == "" {
		return nil, errors.New("networkName is null")
	}

	resp, err := r.networkClient.GetNetwork(
		context.Background(),
		&types.GetNetworkRequest{
			Name: networkName,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("NetworkProvider get network %s failed: %v", networkName, err)
		return nil, err
	}

	return resp.Network, nil
}

// Get network by networkID
func (r *NeutronProvider) GetNetworkByID(networkID string) (*types.Network, error) {
	if networkID == "" {
		return nil, errors.New("networkID is null")
	}

	resp, err := r.networkClient.GetNetwork(
		context.Background(),
		&types.GetNetworkRequest{
			Id: networkID,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("NetworkProvider get network %s failed: %v", networkID, err)
		return nil, err
	}

	return resp.Network, nil
}

// Create network
func (r *NeutronProvider) CreateNetwork(network *types.Network) error {
	resp, err := r.networkClient.CreateNetwork(
		context.Background(),
		&types.CreateNetworkRequest{
			Network: network,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("NetworkProvider create network %s failed: %v", network.Name, err)
		return err
	}

	return nil
}

// Update network
func (r *NeutronProvider) UpdateNetwork(network *types.Network) error {
	resp, err := r.networkClient.UpdateNetwork(
		context.Background(),
		&types.UpdateNetworkRequest{
			Network: network,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("NetworkProvider update network %s failed: %v", network.Name, err)
		return err
	}

	return nil
}

// Delete network by networkName
func (r *NeutronProvider) DeleteNetwork(networkID string) error {
	if networkID == "" {
		return errors.New("networkName is null")
	}

	resp, err := r.networkClient.DeleteNetwork(
		context.Background(),
		&types.DeleteNetworkRequest{
			NetworkID: networkID,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("NetworkProvider delete network %s failed: %v", networkID, err)
		return err
	}

	return nil
}

//List all subnets in the  network
func (r *NeutronProvider) ListSubnets(networkID string) ([]*types.Subnet, error) {
	resp, err := r.subnetClient.ListSubnets(
		context.Background(),
		&types.ListSubnetsRequest{
			NetworkID: networkID,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("NetworkProvider list network %s 's subnets failed: %v", networkID, err)
		return nil, err
	}

	return resp.Subnets, nil
}

func (r *NeutronProvider) CreateSubnet(subnet *types.Subnet) error {
	resp, err := r.subnetClient.CreateSubnet(
		context.Background(),
		&types.CreateSubnetRequest{
			Subnet: subnet,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("NetworkProvider create subnet %s  failed: %v", subnet.Name, err)
		return err
	}

	return nil
}

// Delete a subnet from a network
func (r *NeutronProvider) DeleteSubnet(subnetID string, networkID string) error {
	resp, err := r.subnetClient.DeleteSubnet(
		context.Background(),
		&types.DeleteSubnetRequest{
			SubnetID:  subnetID,
			NetworkID: networkID,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("NetworkProvider delete subnet %s  failed: %v", subnetID, err)
		return err
	}

	return nil
}

//Update subnet
func (r *NeutronProvider) UpdateSubnet(subnet *types.Subnet) error {
	resp, err := r.subnetClient.UpdateSubnet(
		context.Background(),
		&types.UpdateSubnetRequest{
			Subnet: subnet,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("NetworkProvider update subnet %s  failed: %v", subnet.Name, err)
		return err
	}

	return nil
}

// Setup pod
func (r *NeutronProvider) SetupPod(podName, namespace, podInfraContainerID string, network *types.Network, containerRuntime, subnetID string) error {
	resp, err := r.podClient.SetupPod(
		context.Background(),
		&types.SetupPodRequest{
			PodName:             podName,
			Namespace:           namespace,
			PodInfraContainerID: podInfraContainerID,
			ContainerRuntime:    containerRuntime,
			Network:             network,
			SubnetID:            subnetID,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("NetworkProvider SetupPod %s failed: %v", podName, err)
		return err
	}

	return nil
}

// Teardown pod
func (r *NeutronProvider) TeardownPod(podName, namespace, podInfraContainerID string, network *types.Network, containerRuntime string) error {
	resp, err := r.podClient.TeardownPod(
		context.Background(),
		&types.TeardownPodRequest{
			PodName:             podName,
			Namespace:           namespace,
			PodInfraContainerID: podInfraContainerID,
			ContainerRuntime:    containerRuntime,
			Network:             network,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("NetworkProvider TeardownPod %s failed: %v", podName, err)
		return err
	}

	return nil
}

// Status of pod
func (r *NeutronProvider) PodStatus(podName, namespace, podInfraContainerID string, network *types.Network, containerRuntime string) (string, error) {
	resp, err := r.podClient.PodStatus(
		context.Background(),
		&types.PodStatusRequest{
			PodName:             podName,
			Namespace:           namespace,
			PodInfraContainerID: podInfraContainerID,
			ContainerRuntime:    containerRuntime,
			Network:             network,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("NetworkProvider TeardownPod %s failed: %v", podName, err)
		return "", err
	}

	return resp.Ip, nil
}

func main() {
	neutron, err := InitNeutronProviders("127.0.0.1:4237")
	if err != nil {
		fmt.Errorf("init client fail")
	}
	networkName := "br-ex"
	getNetworkResponse, err := neutron.Networks().GetNetwork(networkName)
	if err != nil {
		glog.Errorf("NetworkProvider get network %s failed: %v", networkName, err)
		return
	}
	fmt.Println("%v", getNetworkResponse)

	testNetName := "testnet2"
	/*var subnets []*types.Subnet
	subnet := &types.Subnet{
		Name:    "subnet1",
		Cidr:    "192.168.1.0/24",
		Gateway: "192.168.1.1",
	}
	//subnets = append(subnets, subnet)
	network := &types.Network{
		Name:    testNetName,
		Subnets: subnets,
	}

	err = neutron.Networks().CreateNetwork(network)
	if err != nil {
		glog.Errorf("NetworkProvider create network failed: ", err)
		return
	}
	/*defer func() {
		err = neutron.Networks().DeleteNetwork(testNetName)
		if err != nil {
			glog.Errorf("NetworkProvider delete network filed: ", err)
		}
	}()*/

	getNetworkResponse, err = neutron.Networks().GetNetwork(testNetName)
	if err != nil {
		glog.Errorf("NetworkProvider get network failed: ", err)
		return
	}
	fmt.Println("%v", getNetworkResponse)
	network := &types.Network{
		Uid:  "c2f383a7-1c80-4f71-b987-39214b1597a2",
		Name: "testnet02",
	}
	err = neutron.Networks().UpdateNetwork(network)
	if err != nil {
		glog.Errorf("NetworkProvider update network failed: ", err)
		return
	}

	/*listSubnetResponse, err := neutron.Subnets().ListSubnets("c084cb41-cf08-4d19-abb7-64b3d112baf2")
	if err != nil {
		glog.Errorf("NetworkProvider list subnets failed: ", err)
		return
	}
	fmt.Println("%v", listSubnetResponse)*/
	// gateway can not update
	/*subnet := &types.Subnet{
		NetworkID: "c2f383a7-1c80-4f71-b987-39214b1597a2",
		Name:      "subnet02",
		Cidr:      "192.168.6.0/24",
		Gateway:   "192.168.6.1",
	}
	err = neutron.Subnets().CreateSubnet(subnet)
	if err != nil {
		glog.Errorf("NetworkProvider create subnets failed: ", err)
		return
	}*/
	/*subnet := &types.Subnet{
		Uid:  "1fcbfbc3-fd75-49e6-a981-d5c0ba595d7b",
		Name: "subnet03",
	}
	err = neutron.Subnets().UpdateSubnet(subnet)
	if err != nil {
		glog.Errorf("NetworkProvider create subnets failed: ", err)
		return
	}*/
	/*err = neutron.Subnets().DeleteSubnet("141b5f1b-e2e4-4e89-a8d7-75f7758acb4d", "c084cb41-cf08-4d19-abb7-64b3d112baf2")
	if err != nil {
		glog.Errorf("NetworkProvider delete subnets failed: ", err)
		return
	}

	/*err = neutron.Pods().SetupPod("testPodName1", "testNamespace", "37d7676c80c3", getNetworkResponse, "docker")
	if err != nil {
		glog.Errorf("NetworkProvier create setup pod failed:%v", err)
		return
	}*/

	/*err = neutron.Pods().SetupPod("testPodName3", "testNamespace3", "828db65632e0", getNetworkResponse, "docker")
	if err != nil {
		glog.Errorf("NetworkProvier create setup pod failed:%v", err)
		return
	}*/

}
