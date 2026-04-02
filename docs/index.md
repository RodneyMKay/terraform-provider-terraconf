---
page_title: "Provider: Terraconf"
description: |-
  The Terraconf provider is used to load data from configuration in various 
  formats (currently only YAML) and make it available for interpretation in 
  Terraform. The goal is to improve error reporting for the configuration.
---

# Terraconf Provider

The Terraconf provider is used to load data from configuration in various 
formats (currently only YAML) and make it available for interpretation in 
Terraform. The goal is to improve error reporting for the configuration.

## Example Usage

In this example, we illustrate how one could use the provider to create
Azure Resource Groups based on YAML configuration files. The provider will read
the YAML files, validate them against a JSON schema, and then make the data
available for use in Terraform resources. One of the input files contains a
slight error to demonstrate the provider's error reporting capabilities.

### Terraform

```terraform
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
```

### Input Files

`./data/networking.yml`
```yaml
- name: rg-networking-weu
  location: West Europe
```

`./data/website/frontend.yml`
```yaml
- name: rg-website-frontend
  location: West Europe
```

`./data/website/backend.yml`
```yaml
- name: rg-website-order-service
  location: West Europe

- name: rg-website-customer-service
  location: West Europe

- name: rg-website-database
  location: West Euope
```

`./schema/data.schema.json`
```json
{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "type": "array",
    "items": {
        "type": "object",
        "required": [
            "name",
            "location"
        ],
        "additionalProperties": false,
        "properties": {
            "name": {
                "type": "string"
            },
            "location": {
                "enum": [
                    "West Europe",
                    "Germany West Central"
                ]
            }
        }
    }
}
```

### Plan Results

Running this example produces the following plan:
```txt
Planning failed. Terraform encountered an error while generating this plan.

╷
│ Error: YAML Schema Validation Error
│ 
│   with data.terraconf_yaml.config,
│   on provider.tf line 4, in data "terraconf_yaml" "config":
│    4: data "terraconf_yaml" "config" {
│ 
│ ERROR: In data/website/backend.yml:8:13
│   7 │ - name: rg-website-database
│   8 │   location: West Euope
│     │             ^ value must be one of 'West Europe', 'Germany West Central'
│   9 │ 
│ 
╵
```

## Schema

All settings are configured on the data sources
