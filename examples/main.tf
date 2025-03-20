terraform {
  required_providers {
    aztools = {
      source = "glueckkanja/aztools"
#       version = "0.7.0"
    }
  }
}

locals {
  separator   = "-"
  environment = "dev"
}


provider "aztools" {
  environment        = local.environment
  separator          = local.separator
  convention         = "default"
  lowercase          = true
  schema_naming_path = "./naming_schema/schema.naming.json"
  # schema_naming_url = "https://raw.githubusercontent.com/glueckkanja/terraform-provider-aztools/main/examples/naming_schema/schema.naming.json"
  schema_locations_path = "./naming_schema/schema.locations.json"
  # schema_locations_url = "https://raw.githubusercontent.com/glueckkanja/terraform-provider-aztools/main/examples/naming_schema/schema.locations.json"
}
