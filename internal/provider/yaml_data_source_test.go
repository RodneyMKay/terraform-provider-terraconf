package provider

// import (
// 	"testing"

// 	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
// 	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
// 	"github.com/hashicorp/terraform-plugin-testing/statecheck"
// 	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
// )

// func TestAccExampleDataSource(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			// Read testing
// 			{
// 				Config: testAccExampleDataSourceConfig,
// 				ConfigStateChecks: []statecheck.StateCheck{
// 					statecheck.ExpectKnownValue(
// 						"data.terraconf_yaml.test",
// 						tfjsonpath.New("output"),
// 						knownvalue.StringExact("example-id"),
// 					),
// 				},
// 			},
// 		},
// 	})
// }

// const testAccExampleDataSourceConfig = `
// data "terraconf_yaml" "test" {
// 	input_glob = "./testdata/*.yaml"
// 	json_schema = <<-EOT
// }
// `
