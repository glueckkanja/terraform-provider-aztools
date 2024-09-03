Terraform Provider AzTools
==================

Provider to support implementing [Recommended naming and tagging conventions](https://docs.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/naming-and-tagging) for Azure resources.

Azure defines [naming rules and restrictions](https://docs.microsoft.com/en-us/azure/azure-resource-manager/management/resource-name-rules) for Azure resources. This guidance provides detailed recommendations to support enterprise cloud adoption efforts.

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.5

Using the provider
----------------------

See [Terraform Registry](https://registry.terraform.io/providers/glueckkanja/aztools/latest).

Building The Provider
---------------------

1. Clone the repository
1. Install 'Remote Containers' VS Code Extension
1. Build the provider using the `make` command:
```sh
$ make build
```

Developing the Provider
---------------------------

1. Clone the repository
1. Install 'Remote Containers' VS Code Extension
1. Build & test the provider using the `make` command:
```sh
$ make
```