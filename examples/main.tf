terraform {
  required_providers {
    //azurerm = {
    //  source = "hashicorp/azurerm"
    //}
    aztools = {
      source = "github.com/glueckkanja/aztools"
    }
  }
}

locals {
  sep = "-"
}

//rovider "azurerm" {
// features {}
//

provider "aztools" {
  environment           = "prd"
  separator             = local.sep
  convention            = "default"
  lowercase             = true
  schema_naming_path    = "./naming_schema/schema.naming.json"
  schema_locations_path = "./naming_schema/schema.locations.json"
}
