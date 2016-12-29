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

	//FloatingIP reuturns a FloatingIP interface
	FloatingIPs() FloatingIPs
}

type FloatingIPs interface {
	CreateFloatingIp(tenantID string) (*types.FloatingIp, error)
	BindFloatingIp(protId, floatingipId string) error
	DelFloatingIp(floatingipId string) error
	BindPortToExternal(portName, tenantID string) (string, error)
	UnbindPortFromExternal(portName string) error
	ListFloatingIps(floatingNetworkID string) ([]*types.FloatingIp, error)
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
	// Delete network by networkID
	DeleteNetwork(networkID string) error
	// List network by tenantID
	ListNetworks(tenantID string) ([]*types.Network, error)
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
	floatingIPClient   types.FloatingIPsClient
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
	floatingIPClient := types.NewFloatingIPsClient(conn)
	//	Client(conn)
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
		floatingIPClient:   floatingIPClient,
	}, nil
}

// FloatingIp interface is self
func (r *NeutronProvider) FloatingIPs() FloatingIPs {

	return r
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

//Create FloatingIp
func (r *NeutronProvider) CreateFloatingIp(tenantId string) (*types.FloatingIp, error) {
	resp, err := r.floatingIPClient.CreateFloatingIp(
		context.Background(),
		&types.CreateFloatingIpRequest{
			TenantID: tenantId,
		},
	)

	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("Floating Ip create failed: %v", err)
		return nil, err
	}

	return resp.FloatingIp, nil
}

func (r NeutronProvider) BindFloatingIp(protId, floatingipId string) error {
	if protId == "" || floatingipId == "" {
		glog.Warningf("PortId or FloatingIp Id cant not be nil ")
		return errors.New("PortId or FloatingIp Id cant not be nil")
	}

	resp, err := r.floatingIPClient.BindFloatingIp(
		context.Background(),
		&types.BindFloatingIpRequest{
			PortId:       protId,
			FloatingipId: floatingipId,
		},
	)

	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}

		glog.Warningf("Bind floatingIp: %v to Port: %v failed: %v", floatingipId, protId, err)
		return err
	}

	return nil
}

func (r *NeutronProvider) BindPortToExternal(portName, tenantID string) (string, error) {
	resp, err := r.floatingIPClient.BindPortToExternal(
		context.Background(),
		&types.BindPortToExternalRequest{
			PortName: portName,
			TenantID: tenantID,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("Bind Port: %v to Network failed: %v", portName, err)
		return "", err
	}
	return resp.Floatingip, nil
}

func (r *NeutronProvider) UnbindPortFromExternal(portName string) error {
	resp, err := r.floatingIPClient.UnbindPortFromExternal(
		context.Background(),
		&types.UnbindPortFromExternalRequest{
			PortName: portName,
		},
	)

	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("Unbing Port: %v from Network failed: %v", portName, err)

		return err
	}

	return nil
}

//Del DelFloatingIp by id
func (r *NeutronProvider) DelFloatingIp(floatingipId string) error {
	if floatingipId == "" {
		return errors.New("floatingipId is null")
	}

	resp, err := r.floatingIPClient.DelFloatingIp(
		context.Background(),
		&types.DelFloatingIpRequest{
			FloatingipId: floatingipId,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warning("Deleting floatingIp %s failed : %v", floatingipId, err)
		return err
	}

	return nil
}

func (r *NeutronProvider) ListFloatingIps(floatingNetworkID string) ([]*types.FloatingIp, error) {
	resp, err := r.floatingIPClient.ListFloatingIps(
		context.Background(),
		&types.ListFloatingIpsRequest{
			FloatingNetworkID: floatingNetworkID,
		},
	)
	if err != nil || resp.Error != "" {
		if err != nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("List Floating Ips failed: ", err)
		return nil, err
	}

	return resp.Floatings, nil
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

//List all Networks by tenantID
func (r *NeutronProvider) ListNetworks(tenantID string) ([]*types.Network, error) {
	resp, err := r.networkClient.ListNetworks(
		context.Background(),
		&types.ListNetworkRequest{
			TenantID: tenantID,
		},
	)
	if err != nil || resp.Error != "" {
		if err == nil {
			err = errors.New(resp.Error)
		}
		glog.Warningf("NetworkProvider list networks by tenantID %s failed: %v", tenantID, err)
		return nil, err
	}

	return resp.Networks, nil
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
	neutron, err := InitNeutronProviders("10.10.101.79:4237")
	if err != nil {
		fmt.Errorf("init client fail")
	}

	//	Create a FloatingIP
	//	tenantID := "ae3917b28044403abf8ae7d52c44d2fd"
	//	createFltIpResp, err := neutron.FloatingIPs().CreateFloatingIp(tenantID)
	//	if err != nil {
	//		glog.Errorf("Create FloatingIp failed: %v ", err)
	//		return
	//	}
	//	fmt.Printf("Floating Ip is: %v ", createFltIpResp.FloatingIpAddress)

	//Delete a Floating Ip, if delete succeed,return nil, else return err
	//c4f464eb-5fde-4ba1-9f61-28660ce6c069
	//	floatingipId := "c4f464eb-5fde-4ba1-9f61-28660ce6c069"
	//	delResp := neutron.FloatingIPs().DelFloatingIp(floatingipId)
	//	if err != nil {
	//		glog.Errorf("Delete floatingIP %s failed: %v", floatingipId, err)
	//		return err
	//	}
	//	fmt.Println("=====> %v", delResp)

	// BindPortToExternal, if bind success return binded floating ip, else return error
	//	porNname := "kube_nginx-6xgdn_zzc_6b815cb8-e68f-41ea-a927-444e650a568"
	//	// tenantID is port's tenantID
	//	tenantID := "ae3917b28044403abf8ae7d52c44d2fd"
	//	bindResp, err := neutron.FloatingIPs().BindPortToExternal(porNname, tenantID)
	//	if err != nil {
	//		glog.Errorf("Bind failed %v", err)
	//	}
	//	fmt.Println(bindResp)

	//	//UnbindPortFromExternal
	//	protName := "kube_nginx-6xgdn_zzc_6b815cb8-e68f-41ea-a927-444e650a568f"
	//	resp := neutron.FloatingIPs().UnbindPortFromExternal(protName)
	//	if err != nil {
	//		glog.Errorf("Bind failed %v", err)
	//	}
	//	fmt.Println(resp)

	// ListFloatingIps, if find return an array ,if not find return a nil array []
	//	netId := "5378ca69-6588-4516-97fd-fcb9a357cc16"
	//	resp, err := neutron.FloatingIPs().ListFloatingIps(netId)
	//	if err != nil {
	//		glog.Errorf("List FLI failed : %v", err)
	//		return
	//	}
	//	fmt.Println(resp)

	//BindFloatingIp, Bind success return nil, else return err
	portid := "6f90e2e0-4ba9-47be-bb79-bf2fb140d62f"
	fltid := "09bc5e1a-cd66-4e57-9ef0-d9f3dc3f500"
	resp := neutron.FloatingIPs().BindFloatingIp(portid, fltid)
	if resp != nil {
		glog.Errorf("Bind failed: %v", resp)
		return
	}
	fmt.Printf("Error:", resp)

	//	networkName := "public"
	//	getNetworkResponse, err := neutron.Networks().GetNetwork(networkName)
	//	if err != nil {
	//		glog.Errorf("NetworkProvider get network %s failed: %v", networkName, err)
	//		return
	//	}
	//	fmt.Println("%v", getNetworkResponse)

	//  testNetName := "testnet4"
	/*var subnets []*types.Subnet
	subnet := &types.Subnet{
		Name:    "subnet1",
		Cidr:    "192.168.11.0/24",
		Gateway: "192.168.11.1",
	}
	subnets = append(subnets, subnet)
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

	//	getNetworkResponse, err = neutron.Networks().GetNetwork(testNetName)
	//	if err != nil {
	//		glog.Errorf("NetworkProvider get network failed: ", err)
	//		return
	//	}
	//	fmt.Println("%v", getNetworkResponse)

	/*network := &types.Network{
		Uid:  "c2f383a7-1c80-4f71-b987-39214b1597a2",
		Name: "testnet02",
	}
	err = neutron.Networks().UpdateNetwork(network)
	if err != nil {
		glog.Errorf("NetworkProvider update network failed: ", err)
		return
	}*/

	/*ListNetworkResponse, err := neutron.Networks().ListNetworks("414454afc37f4ff395d06e61167f8108")
	if err != nil {
		glog.Errorf("NetworkProvider list Networks failed: ", err)
		return
	}
	fmt.Println("%v", ListNetworkResponse)*/

	/*err = neutron.Networks().DeleteNetwork("488cfc23-0852-4853-bf1b-f17d341cde10")
	fmt.Println("%v", err)
	if err != nil {
		glog.Errorf("NetworkProvider delete network failed: ", err)
		return
	}

	/*listSubnetResponse, err := neutron.Subnets().ListSubnets("c084cb41-cf08-4d19-abb7-64b3d112baf2")
	if err != nil {
		glog.Errorf("NetworkProvider list subnets failed: ", err)
		return
	}
	fmt.Println("%v", listSubnetResponse)*/

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
	}

	// gateway can not update
	/*subnet := &types.Subnet{
		Uid:  "1fcbfbc3-fd75-49e6-a981-d5c0ba595d7b",
		Name: "subnet03",
	}
	err = neutron.Subnets().UpdateSubnet(subnet)
	if err != nil {
		glog.Errorf("NetworkProvider create subnets failed: ", err)
		return
	}*/
	/*err = neutron.Subnets().DeleteSubnet("1fcbfbc3-fd75-49e6-a981-d5c0ba595d7b", "c2f383a7-1c80-4f71-b987-39214b1597a2")
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
