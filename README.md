# GoLang Ansible Inventory Module for OpenStack

## Description

GoLang Ansible Inventory Module for OpenStack aim to provide Ansible users a simple way to generate dynamic inventory file of OpenStack Cloud resources, based on hosts Metadata.

## Basic Usage

Run:

```bash
go build -o ansible-openstak
```

to generate an executable of this module. Copy the **ansible-openstack** bin into the root of your ansible project.

To connect to an OpenStack instance, two more files are required in the Ansible project root:

1. [clouds.yaml](#clouds.yaml) (alredy provided)
1. [secure.yaml](#secure.yaml) (need to be created manually)

### clouds.yaml

```yaml
# clouds.yaml
clouds:
  ocloud:
    auth:
      username: "this should be overrided by secure.yaml"
      password: "this should be overrided by secure.yaml"
      project_name: "this should be overrided by secure.yaml"
      auth_url: https://api.it-mil1.entercloudsuite.com/v2.0
    regions:
      - it-mil1
```

Customize it following your needs.

### secure.yaml

```yaml
clouds:
  ocloud:
    auth:
      username: "<Your username>"
      password: "<Your password>"
      project_name: "<Your target Project name>"
    auth_type: "password"
```

Customize it following your needs.

Run:

```bash
sudo ansible-openstack [tag list]
```

to run the application with default parameters and generate a default inventory in hosts/inventory.ini

**Note: hosts Metadata must be prior generated and binded to the hosts (manually or by IaC tools)**

## Help

To obtain more informations about usage, run:

```bash
./ansible-openstack -h
```

This will produce the following help screen:

```bash
Usage of ./ansible-openstack:
  -domain string
        Specify domain name for the hosts (default ".it-mil1.ecs.compute.internal")
  -filename string
        Specify inventory output filename (default "inventory.ini")
  -output string
        Specify output path for inventory.ini file (default "./hosts")
```
