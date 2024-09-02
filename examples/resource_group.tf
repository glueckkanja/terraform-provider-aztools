resource "aztools_resource_name" "rg" {
  resource_type = "azurerm_resource_group"
  name          = "example_rg"
  location      = "westeurope"
  prefixes      = ["rg"]
  suffixes      = ["test", "001"]
}

resource "azurerm_resource_group" "rg" {
  location = "westeurope"
  name     = aztools_resource_name.rg.result
}