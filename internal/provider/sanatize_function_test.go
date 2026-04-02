package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestAccSanatizeFunction(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSanatizeFunctionConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"sanatized",
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"foo": knownvalue.StringExact("Hello"),
							"bar": knownvalue.StringExact("World"),
							"baz": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"foobar": knownvalue.StringExact("1"),
								}),
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"foobar": knownvalue.StringExact("2"),
								}),
							}),
						}),
					),
				},
			},
		},
	})
}

const testAccSanatizeFunctionConfig = `
locals {
  input = {
    _terraconf = "special tag meant to be removed"
    foo        = "Hello"
    bar        = "World"

    baz = [
	  {
        _terraconf = "special tag meant to be removed"
        foobar     = "1"
      },
	  {
        _terraconf = "special tag meant to be removed"
        foobar     = "2"
      }
    ]
  }
}

output "sanatized" {
  value = provider::terraconf::sanatize(local.input)
}
`
