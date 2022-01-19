package main

import (
	"io/ioutil"
	"log"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
	yaml "gopkg.in/yaml.v3"
)

func main() {

	//TODO: retrive cli args to parametrize search tags

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

	// TODO: call the following functions for every tag passed as args
	ids, names, err := retriveServers(client, server_opts)
	if err != nil {
		log.Panicf("Could not retrive servers: %v\n", err)
		return
	}

	ips, err := retriveNetworAddress(client, ids)
	if err != nil {
		log.Panicf("Could not retrive network addresses: %v\n", err)
		return
	}

	//TODO: make a map merge based on tag list
	results := composeInventoryMap(names, ips)
	log.Printf("Results: %v\n", results)

	toYaml, err := composeInventory(&results)
	if err != nil {
		log.Panicf("Could not convert to yaml: %v\n", err)
		return
	}

	err = generateInventoryFile(toYaml, "inventory.yml")
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

func retriveServers(client *gophercloud.ServiceClient, server_opts servers.ListOpts) (ids []string, names []string, err error) {
	pager := servers.List(client, server_opts)
	err = pager.EachPage(func(p pagination.Page) (bool, error) {
		serverlist, err := servers.ExtractServers(p)
		for _, v := range serverlist {
			for _, v2 := range v.Metadata {
				if v2 == "node" || v2 == "master" {
					ids = append(ids, v.ID)
					names = append(names, v.Name)
				}
			}
		}
		return false, err
	})
	if err != nil {
		return nil, nil, err
	}
	return ids, names, nil
}

func retriveNetworAddress(client *gophercloud.ServiceClient, ids []string) (ips []string, err error) {
	for _, id := range ids {
		pager := servers.ListAddresses(client, id)
		err = pager.EachPage(func(p pagination.Page) (bool, error) {
			addressList, err := servers.ExtractAddresses(p)
			for _, addresses := range addressList {
				for idx := 1; idx < len(addresses); idx += 2 {
					if addresses[idx].Version == 4 {
						ips = append(ips, addresses[idx].Address)
					}
				}
			}
			return false, err
		})
	}
	if err != nil {
		return nil, err
	}
	return ips, nil
}

func composeInventoryMap(ids []string, ips []string) (results map[string]string) {
	results = map[string]string{}
	if len(ids) == len(ips) {
		for i := range ids {
			results[ids[i]] = ips[i]
		}
	}
	return results
}

func composeInventory(hosts *map[string]string) (output []byte, err error) {
	output, err = yaml.Marshal(*hosts)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func generateInventoryFile(content []byte, path string) (err error) {
	ioutil.WriteFile(path, content, 0655)
	if err != nil {
		return err
	}
	return nil
}
