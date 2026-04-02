provider "terraconf" {
}

data "terraconf_yaml" "config" {
  input_glob  = "./data/**/*.{yaml,yml}"
  schema_file = "./schema/data.schema.json"
}

resource "azurerm_resource_group" "resource_groups" {
  for_each = {
    for group in data.terraconf_yaml.config.output
    : group.name => group
  }

  name     = each.key
  location = each.value.location
}
