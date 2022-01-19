package main

import (
	"log"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
	yaml "gopkg.in/yaml.v3"
)

type Inventory struct {
	Tag     string    `yaml:"group"`
	Element []Element `yaml:"nodes"`
}

type Element struct {
	Id   string `yaml:"id"`
	Name string `yaml:"hostname"`
	Ip   string `yaml:"ip"`
}

var fullInventory []Inventory

func main() {

	args := os.Args

	opts := gophercloud.AuthOptions{
		IdentityEndpoint: "https://api.it-mil1.entercloudsuite.com/v2.0",
		Username:         "andrea.colaiuda@irideos.it",
		Password:         "Gn0m0@i986!_",
		TenantName:       "andrea.colaiuda@irideos.it",
		Scope:            &gophercloud.AuthScope{ProjectName: "andrea.colaiuda@irideos.it"},
	}

	server_opts := servers.ListOpts{}

	provider, err := initProvider(opts)
	if err != nil {
		log.Fatalf("Error initializing Openstack provider: %v\n", err)
		return
	}

	client, err := initCompute(provider)
	if err != nil {
		log.Fatalf("Error initializing Compute client: %v\n", err)
		return
	}

	for tagIdx := 1; tagIdx < len(args); tagIdx++ {
		inventory, err := retriveServers(client, server_opts, args[tagIdx])
		if err != nil {
			log.Panicf("Could not retrive servers: %v\n", err)
			return
		}
		fullInventory = append(fullInventory, inventory)
	}

	err = generateInventoryFile(fullInventory, "inventory.yml")
	if err != nil {
		log.Panicf("Error writing YAML file: %v\n", err)
		return
	}

}

func initProvider(opts gophercloud.AuthOptions) (provider *gophercloud.ProviderClient, err error) {
	provider, err = openstack.AuthenticatedClient(opts)
	if err != nil {
		log.Fatalf("Error during authentication: %v", err)
	}
	if err != nil {
		return nil, err
	}
	return provider, nil
}

func initCompute(provider *gophercloud.ProviderClient) (client *gophercloud.ServiceClient, err error) {
	client, err = openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: "it-mil1",
		Type:   "compute",
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func retriveServers(client *gophercloud.ServiceClient, server_opts servers.ListOpts, tag string) (inventory Inventory, err error) {
	pager := servers.List(client, server_opts)
	pager.EachPage(func(p pagination.Page) (bool, error) {
		serverlist, err := servers.ExtractServers(p)
		for _, v := range serverlist {
			for _, v2 := range v.Metadata {
				if v2 == tag {
					var element Element
					element.Id = v.ID
					element.Name = v.Name
					element.Ip, err = retriveNetworAddress(client, element.Id)
					inventory.Element = append(inventory.Element, element)
					inventory.Tag = tag
				}
			}
		}
		return false, err
	})
	return inventory, nil
}

func retriveNetworAddress(client *gophercloud.ServiceClient, id string) (ip string, err error) {
	pager := servers.ListAddresses(client, id)
	err = pager.EachPage(func(p pagination.Page) (bool, error) {
		addressList, err := servers.ExtractAddresses(p)
		for _, addresses := range addressList {
			for idx := 1; idx < len(addresses); idx += 2 {
				if addresses[idx].Version == 4 {
					ip = addresses[idx].Address
				}
			}
		}
		return false, err
	})
	if err != nil {
		return "", err
	}
	return ip, nil
}

func generateInventoryFile(hosts []Inventory, path string) (err error) {
	output, err := yaml.Marshal(hosts)
	if err != nil {
		return err
	}
	os.WriteFile(path, output, 0644)
	if err != nil {
		return err
	}
	return nil
}
