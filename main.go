package main

import (
	"flag"
	"log"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/txn2/txeh"
	ini "gopkg.in/ini.v1"
)

type Inventory struct {
	tag     string
	Element []Element
}

type Element struct {
	id   string
	Name string
	Ip   string
}

var fullInventory []Inventory

func main() {

	path := flag.String("output", "./hosts", "Specify output path for inventory.ini file")
	filename := flag.String("filename", "inventory.ini", "Specify inventory output filename")
	domainName := flag.String("domain", ".it-mil1.ecs.compute.internal", "Specify domain name for the hosts")

	flag.Parse()

	args := os.Args

	server_opts := servers.ListOpts{}

	if _, err := os.Stat(*path); err != nil {
		if err := os.Mkdir(*path, os.ModePerm); err != nil {
			log.Fatalf("Erorr: %v\n", err)
		}
	}
	os.Remove(*path + "/" + *filename)
	os.Create(*path + "/" + *filename)

	provider, err := initProvider()
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
		inventory, err := retriveServers(client, server_opts, args[tagIdx], *domainName)
		if err != nil {
			log.Panicf("Could not retrive servers: %v\n", err)
			return
		}
		fullInventory = append(fullInventory, inventory)
	}

	err = generateInventoryFile(fullInventory, *path+"/"+*filename, *domainName)
	if err != nil {
		log.Panicf("Error writing INI file: %v\n", err)
		return
	}

}

func authenticate() (client *gophercloud.ProviderClient, err error) {
	opts := &clientconfig.ClientOpts{
		Cloud: "ocloud",
	}
	client, err = clientconfig.AuthenticatedClient(opts)
	if err != nil {
		log.Fatalf("Error authenticating user: %v\n", err)
		return nil, err
	}
	return client, nil
}

func initProvider() (provider *gophercloud.ProviderClient, err error) {
	provider, err = authenticate()
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

func retriveServers(client *gophercloud.ServiceClient, server_opts servers.ListOpts, tag string, domain string) (inventory Inventory, err error) {
	pager := servers.List(client, server_opts)
	pager.EachPage(func(p pagination.Page) (bool, error) {
		serverlist, err := servers.ExtractServers(p)
		for _, v := range serverlist {
			for _, v2 := range v.Metadata {
				if v2 == tag {
					var element Element
					element.id = v.ID
					element.Name = v.Name
					element.Ip, err = retriveNetworAddress(client, element.id)
					inventory.Element = append(inventory.Element, element)
					inventory.tag = tag
					editHostsFile(element.Ip, element.Name+domain)
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
				// Get only IPv4
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

func generateInventoryFile(hosts []Inventory, path string, domain string) (err error) {
	output, err := ini.LoadSources(ini.LoadOptions{
		AllowBooleanKeys: true,
	}, path)
	if err != nil {
		return err
	}
	for _, host := range hosts {
		output.Section("staging:children").NewBooleanKey(host.tag)
		for _, element := range host.Element {
			output.Section(host.tag).NewBooleanKey(element.Name + domain)
		}
	}

	output.SaveTo(path)
	if err != nil {
		return err
	}
	return nil
}

func editHostsFile(ip string, hostname string) {
	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	hosts.AddHost(ip, hostname)
	if err := hosts.Save(); err != nil {
		log.Fatalf("Error saving hosts file: %v\n", err)
	}

}
