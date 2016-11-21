/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubestack

import (
	"net"

	"github.com/golang/glog"
	"github.com/hyperhq/kubestack/pkg/common"
	provider "github.com/hyperhq/kubestack/pkg/types"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// KubeHandler forwards requests and responses between the docker daemon and the plugin.
type KubeHandler struct {
	driver *common.OpenStack
	server *grpc.Server
}

// NewKubeHandler initializes the request handler with a driver implementation.
func NewKubeHandler(driver *common.OpenStack) *KubeHandler {
	h := &KubeHandler{
		driver: driver,
		server: grpc.NewServer(),
	}
	h.registerServer()
	return h
}

func (h *KubeHandler) Serve(addr string) error {
	glog.V(1).Infof("Starting kubestack at %s", addr)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		glog.Fatalf("Failed to listen: %s", addr)
		return err
	}
	return h.server.Serve(l)
}

func (h *KubeHandler) registerServer() {
	provider.RegisterLoadBalancersServer(h.server, h)
	provider.RegisterNetworksServer(h.server, h)
	provider.RegisterPodsServer(h.server, h)
	provider.RegisterSubnetsServer(h.server, h)
}

func (h *KubeHandler) Active(c context.Context, req *provider.ActiveRequest) (*provider.ActivateResponse, error) {
	glog.V(3).Infof("Activating called")

	resp := provider.ActivateResponse{
		Result: true,
	}

	return &resp, nil
}

func (h *KubeHandler) CheckTenantID(c context.Context, req *provider.CheckTenantIDRequest) (*provider.CheckTenantIDResponse, error) {
	glog.V(4).Infof("CheckTenantID with request %v", req.TenantID)

	resp := provider.CheckTenantIDResponse{}
	checkResult, err := h.driver.CheckTenantID(req.TenantID)
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Result = checkResult
	}

	glog.V(4).Infof("CheckTenantID result %v", resp)
	return &resp, nil
}

func (h *KubeHandler) GetNetwork(c context.Context, req *provider.GetNetworkRequest) (*provider.GetNetworkResponse, error) {
	glog.V(4).Infof("GetNetwork with request %v", req.String())

	resp := provider.GetNetworkResponse{}
	var result *provider.Network
	var err error
	if req.Id != "" {
		result, err = h.driver.GetNetworkByID(req.Id)
	} else if req.Name != "" {
		result, err = h.driver.GetNetwork(req.Name)
	}

	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Network = result
	}

	glog.V(4).Infof("GetNetwork result %v", resp)
	return &resp, nil
}

func (h *KubeHandler) CreateNetwork(c context.Context, req *provider.CreateNetworkRequest) (*provider.CommonResponse, error) {
	glog.V(4).Infof("CreateNetwork with request %v", req)

	resp := provider.CommonResponse{}
	req.Network.TenantID = h.driver.ToTenantID(req.Network.TenantID)
	err := h.driver.CreateNetwork(req.Network)
	if err != nil {
		resp.Error = err.Error()
	}

	glog.V(4).Infof("CreateNetwork result %v", resp)
	return &resp, nil
}

func (h *KubeHandler) UpdateNetwork(c context.Context, req *provider.UpdateNetworkRequest) (*provider.CommonResponse, error) {
	glog.V(4).Infof("UpdateNetwork with request %v", req.String())

	resp := provider.CommonResponse{}
	req.Network.TenantID = h.driver.ToTenantID(req.Network.TenantID)
	err := h.driver.UpdateNetwork(req.Network)
	if err != nil {
		resp.Error = err.Error()
	}

	glog.V(4).Infof("UpdateNetwork result %v", resp)
	return &resp, nil
}

func (h *KubeHandler) DeleteNetwork(c context.Context, req *provider.DeleteNetworkRequest) (*provider.CommonResponse, error) {
	glog.V(4).Infof("DeleteNetwork with request %v", req.String())

	resp := provider.CommonResponse{}
	err := h.driver.DeleteNetwork(req.NetworkID)
	if err != nil {
		resp.Error = err.Error()
	}

	glog.V(4).Infof("DeleteNetwork result %v", resp)
	return &resp, nil
}

func (h *KubeHandler) ListNetworks(c context.Context, req *provider.ListNetworkRequest) (*provider.ListNetworkResponse, error) {
	glog.V(4).Infof("ListNetworks with request %v", req.String())

	resp := provider.ListNetworkResponse{}
	var result []*provider.Network
	var err error

	result, err = h.driver.ListNetworks(req.TenantID)

	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Networks = result
	}

	glog.V(4).Infof("ListNetworks result %v", resp)
	return &resp, nil
}

func (h *KubeHandler) ListSubnets(c context.Context, req *provider.ListSubnetsRequest) (*provider.ListSubnetsResponse, error) {
	glog.V(4).Infof("ListSubnets with request %v", req.String())

	resp := provider.ListSubnetsResponse{}
	var result []*provider.Subnet
	var err error

	result, err = h.driver.ListSubnets(req.NetworkID)

	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Subnets = result
	}

	glog.V(4).Infof("ListSubnets result %v", resp)
	return &resp, nil
}

func (h *KubeHandler) CreateSubnet(c context.Context, req *provider.CreateSubnetRequest) (*provider.CommonResponse, error) {
	glog.V(4).Infof("CreateSubnet with request %v", req)

	resp := provider.CommonResponse{}
	req.Subnet.Tenantid = h.driver.ToTenantID(req.Subnet.Tenantid)
	err := h.driver.CreateSubnet(req.Subnet)
	if err != nil {
		resp.Error = err.Error()
	}

	glog.V(4).Infof("CreateSubnet result %v", resp)
	return &resp, nil
}

func (h *KubeHandler) GetSubnet(c context.Context, req *provider.GetSubnetRequest) (*provider.GetSubnetResponse, error) {
	glog.V(4).Infof("GetSubnet with request %v", req.String())

	resp := provider.GetSubnetResponse{}
	var result *provider.Subnet
	var err error
	result, err = h.driver.GetSubnet(req.SubnetID)
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Subnet = result
	}

	glog.V(4).Infof("GetSubnet result %v", resp)
	return &resp, nil
}

func (h *KubeHandler) DeleteSubnet(c context.Context, req *provider.DeleteSubnetRequest) (*provider.CommonResponse, error) {
	glog.V(4).Infof("DeleteSubnet with request %v", req.String())

	resp := provider.CommonResponse{}
	err := h.driver.DeleteSubnet(req.SubnetID, req.NetworkID)
	if err != nil {
		resp.Error = err.Error()
	}

	glog.V(4).Infof("DeleteSubnet result %v", resp)
	return &resp, nil
}

func (h *KubeHandler) UpdateSubnet(c context.Context, req *provider.UpdateSubnetRequest) (*provider.CommonResponse, error) {
	glog.V(4).Infof("UpdateSubnet with request %v", req.String())

	resp := provider.CommonResponse{}
	err := h.driver.UpdateSubnet(req.Subnet)
	if err != nil {
		resp.Error = err.Error()
	}

	glog.V(4).Infof("UpdateSubnet result %v", resp)
	return &resp, nil
}

func (h *KubeHandler) ConnectSubnets(c context.Context, req *provider.ConnectSubnetsRequest) (*provider.CommonResponse, error) {
	glog.V(4).Infof("ConnectSubnets with request %v", req.String())

	resp := provider.CommonResponse{}
	err := h.driver.ConnectSubnets(req.Subnet1, req.Subnet2)
	if err != nil {
		resp.Error = err.Error()
	}

	glog.V(4).Infof("ConnectSubnets result %v", resp)
	return &resp, nil
}

func (h *KubeHandler) GetLoadBalancer(c context.Context, req *provider.GetLoadBalancerRequest) (*provider.GetLoadBalancerResponse, error) {
	resp := provider.GetLoadBalancerResponse{}
	lb, err := h.driver.GetLoadBalancer(req.Name)
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.LoadBalancer = lb
	}

	return &resp, nil
}

func (h *KubeHandler) CreateLoadBalancer(c context.Context, req *provider.CreateLoadBalancerRequest) (*provider.CreateLoadBalancerResponse, error) {
	glog.V(4).Infof("CreateLoadBalancer with request %v", req.String())

	resp := provider.CreateLoadBalancerResponse{}
	req.LoadBalancer.TenantID = h.driver.ToTenantID(req.LoadBalancer.TenantID)
	vip, err := h.driver.CreateLoadBalancer(req.LoadBalancer, string(req.Affinity))
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Vip = vip
	}

	glog.V(4).Infof("CreateLoadBalancer result %v", resp)
	return &resp, nil
}

func (h *KubeHandler) UpdateLoadBalancer(c context.Context, req *provider.UpdateLoadBalancerRequest) (*provider.UpdateLoadBalancerResponse, error) {
	glog.V(4).Infof("UpdateLoadBalancer with request %v", req.String())

	resp := provider.UpdateLoadBalancerResponse{}
	vip, err := h.driver.UpdateLoadBalancer(req.Name, req.Hosts, req.ExternalIPs)
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Vip = vip
	}

	glog.V(4).Infof("UpdateLoadBalancer result %v", resp)

	return &resp, nil
}

func (h *KubeHandler) DeleteLoadBalancer(c context.Context, req *provider.DeleteLoadBalancerRequest) (*provider.CommonResponse, error) {
	glog.V(4).Infof("DeleteLoadBalancer with request %v", req.String())

	resp := provider.CommonResponse{}
	err := h.driver.DeleteLoadBalancer(req.Name)
	if err != nil {
		resp.Error = err.Error()
	}

	glog.V(4).Infof("DeleteLoadBalancer result %v", resp)
	return &resp, nil
}

func (h *KubeHandler) SetupPod(c context.Context, req *provider.SetupPodRequest) (*provider.CommonResponse, error) {
	glog.V(4).Infof("SetupPod with request %v", req.String())

	resp := provider.CommonResponse{}
	// TODO: Add hostname in SetupPod Interface
	err := h.driver.SetupPod(req.PodName, req.Namespace, req.PodInfraContainerID, req.Network, req.ContainerRuntime, req.SubnetID)
	if err != nil {
		glog.Errorf("SetupPod failed: %v", err)
		resp.Error = err.Error()
	}

	return &resp, nil
}

func (h *KubeHandler) TeardownPod(c context.Context, req *provider.TeardownPodRequest) (*provider.CommonResponse, error) {
	glog.V(4).Infof("TeardownPod with request %v", req.String())

	resp := provider.CommonResponse{}
	err := h.driver.TeardownPod(req.PodName, req.Namespace, req.PodInfraContainerID, req.Network, req.ContainerRuntime)
	if err != nil {
		glog.Errorf("TeardownPod failed: %v", err)
		resp.Error = err.Error()
	}

	return &resp, nil
}

func (h *KubeHandler) PodStatus(c context.Context, req *provider.PodStatusRequest) (*provider.PodStatusResponse, error) {
	glog.V(4).Infof("PodStatus with request %v", req.String())

	resp := provider.PodStatusResponse{}
	ip, err := h.driver.PodStatus(req.PodName, req.Namespace, req.PodInfraContainerID, req.Network, req.ContainerRuntime)
	if err != nil {
		glog.Errorf("PodStatus failed: %v", err)
		resp.Error = err.Error()
	} else {
		resp.Ip = ip
	}

	return &resp, nil
}
