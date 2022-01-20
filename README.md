# GoLang Ansible Inventory Module for OpenStack

## Description

GoLang Ansible Inventory Module for OpenStack aim to provide Ansible users a simple way to generate dynamic inventory file of OpenStack Cloud resources, based on hosts Metadata.

## Basic Usage

To run the application do the following steps:

```bash
git clone git@git.kqi.it:gitops/golang/ansible-openstack.git
cd ansible-openstack
go build -o ansible-openstak
```

this will generate a bin file of this application.

Copy the **ansible-openstack** bin into the root of your ansible project.

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
#secure.yaml
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

It will also update your local /etc/hosts file with the public ips and host names retrived from cloud, to allow you to quick reach the hosts with Ansible.

### Note: hosts Metadata must be prior generated and binded to the hosts (manually or by IaC tools)

## Help

To obtain more informations about usage, run:

```bash
./ansible-openstack -h
```

This will produce the following help screen:

```bash
Usage of /tmp/go-build755480796/b001/exe/ansible-openstack:
  -domain string
        Specify domain name for the hosts (default ".it-mil1.ecs.compute.internal")
  -filename string
        Specify inventory output filename (default "inventory.ini")
  -main-group string
        Specify the main group file contained in the inventory file. Use always <name>:children form, otherwise it will break the code! (default "staging:children")
  -output string
        Specify output path for inventory.ini file (default "./hosts")
```
