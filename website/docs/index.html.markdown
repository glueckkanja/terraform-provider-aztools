---
layout: "aztools"
page_title: "Provider: AzTools"
sidebar_current: "docs-aztools-index"
description: |-
  Terraform provider AzTools.
---

# AzTools Provider

Provider to support implementing [Recommended naming and tagging conventions](https://docs.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/naming-and-tagging) for Azure resources.

Azure defines [naming rules and restrictions](https://docs.microsoft.com/en-us/azure/azure-resource-manager/management/resource-name-rules) for Azure resources. This guidance provides detailed recommendations to support enterprise cloud adoption efforts.

Use the navigation to the left to read about the available resources.

~> **Note:** This resource it's in preview and may change in future releases.

## Example Usage

File path

```hcl
provider "aztools" {
  environment           = "dev"
  separator             = "-"
  convention            = "default"
  lowercase             = true
  hash_length           = 2
  schema_naming_path    = "./naming_schema/schema.naming.json"
  schema_locations_path = "./naming_schema/schema.locations.json"
}
```
Url location

```hcl
provider "aztools" {
  environment          = "dev"
  separator            = "-"
  convention           = "default"
  lowercase            = true
  hash_length          = 2
  schema_naming_url    = "https://raw.githubusercontent.com/glueckkanja/terraform-provider-aztools/main/examples/naming_schema/schema.naming.json"
  schema_locations_url = "https://raw.githubusercontent.com/glueckkanja/terraform-provider-aztools/main/examples/naming_schema/schema.locations.json"
}
```

## Argument Reference

The following arguments are supported:

* `environment` - (Optional) Environment atribute. Defaults to `sandbox`.
* `separator` - (Optional) Separator between resource arguments. Defaults to `-`.
* `lowercase` - (Optional) Convert result to lowercase. Possible values are: `true` or `false`. Defaults to `false`.
* `hash_length` - (Optional) Length of random hash. Ovveride all json schema definitions.
* `schema_naming_path` - (Optional) Relative file path from root module to json schema file.
* `schema_locations_path` - (Optional) Relative file path from root module to json schema file.
* `schema_naming_url` - (Optional) Url of json schema file.
* `schema_locations_url` - (Optional) Url of json schema file.


~> **Note:** Parameters defined on resource level will override provider configuration.
~> **Note:** One of `schema_naming_path` or `schema_naming_url` must be specified.
~> **Note:** One of `schema_locations_path` or `schema_locations_url` must be specified.

## Example schema.naming.json

```json
[
  {
    "resourceType": "azurerm_resource_group",
    "abbreviation": "rg",
    "minLength": 1,
    "maxLength": 90,
    "validationRegex": "^[a-zA-Z0-9-._()]{0,89}[a-zA-Z0-9-_()]$",
    "configuration": {
      "useEnvironment": true,
      "useLowerCase": false,
      "useSeparator": true,
      "denyDoubleHyphens": true,
      "namePrecedence": [],
      "hashLength": 4
    }
  }
]
```

Result: `rg-prefixes-example-sandbox-suffixes-abcd-001`

```json
[
  {
    "resourceType": "azurerm_resource_group_custom",
    "abbreviation": "rg",
    "minLength": 1,
    "maxLength": 90,
    "validationRegex": "^[a-zA-Z0-9-._()]{0,89}[a-zA-Z0-9-_()]$",
    "configuration": {
      "useEnvironment": true,
      "useLowerCase": false,
      "useSeparator": true,
      "denyDoubleHyphens": true,
      "namePrecedence": ["prefixes", "name", "location", "environment", "hash", "suffixes", "abbreviation"],
      "hashLength": 4
    }
  }
]
```

Result: `prefixes-example-weeu-sandbox-abcd-suffixes-001-rg`

## Example schema.locations.json

```json
{
  "westeurope": "weeu",
  "norteurope": "neeu"
}
```