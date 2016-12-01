package main

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/portsbinding"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	//"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/fwaas/firewalls"
	//"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/fwaas/policies"
	//"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/fwaas/rules"
	//"github.com/gophercloud/gophercloud/pagination"
)

func main() {
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: "http://10.10.103.70:5000/v2.0",
		Username:         "admin",
		Password:         "admin",
		TenantName:       "admin",
	}
	provider, _ := openstack.AuthenticatedClient(opts)
	client, _ := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{
		Name:   "neutron",
		Region: "RegionOne",
	})

	var fixedIPs [1]ports.IP
	fixedIPs[0] = ports.IP{
		SubnetID: "5f92bdf4-6cb0-47c5-80ae-f2966dd90210",
	}
	opts1 := portsbinding.CreateOpts{
		HostID: "ubuntu",
		//DNSName: "testpod",
		CreateOptsBuilder: ports.CreateOpts{
			NetworkID: "46c55ce1-caa2-4f24-8f3e-805435d13762",
			Name:      "testpod",
			TenantID:  "5c62ef576dc7444cbb73b1fe84b97648",
			//DeviceID:    "123456",
			//FixedIPs:    fixedIPs,
			DeviceOwner: fmt.Sprintf("compute:%s", "ubuntu"),
		},
	}

	port, err := portsbinding.Create(client, opts1).Extract()
	if err != nil {
		fmt.Println("Create port failed: %v", err)
	}
	fmt.Println(port)
	/*options := rules.CreateOpts{
		TenantID:    "5c62ef576dc7444cbb73b1fe84b97648",
		Protocol:    "",
		Description: "123",
		Name:        "any",
		Action:      "allow",
	}

	_, err := rules.Create(client, options).Extract()
	if err != nil {
		fmt.Println(err)
	}*/

	/*policies.List(client, policies.ListOpts{Name: "testnet1"}).EachPage(func(page pagination.Page) (bool, error) {
		actual, err := policies.ExtractPolicies(page)
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		fmt.Println(actual[0])
		return true, nil
	})

	/*rule, _ := rules.Get(client, "e0f981e4-35f2-41f7-a82c-5f5f85cf38a6").Extract()
	//fmt.Println(err)
	fmt.Printf("%v", rule)
	/*options1 := policies.CreateOpts{
		TenantID:    "414454afc37f4ff395d06e61167f8108",
		Name:        "policy",
		Description: "Firewall policy",
		Shared:      gophercloud.Disabled,
		Audited:     gophercloud.Disabled,
		Rules: []string{
			rule.ID,
		},
	}

	osPolicy, err := policies.Create(client, options1).Extract()
	if err != nil {
		fmt.Println(err)
	}
	var routers []string
	routers = append(routers, "67729bbe-95f1-40b7-8745-05cf25ba9b09")
	options := firewalls.CreateOpts{
		TenantID:     "414454afc37f4ff395d06e61167f8108",
		Name:         "fw",
		Description:  "OpenStack firewall",
		AdminStateUp: gophercloud.Enabled,
		PolicyID:     "951583aa-d3bd-4a72-9744-401089c456d1",
		Router_ids:   routers,
	}
	osPolicy.ID
	_, err := firewalls.Create(client, options).Extract()
	if err != nil {
		fmt.Println(err)
	}*/
}
