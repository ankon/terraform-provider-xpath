package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccQueryOneDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccQueryOneDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.xpath_query_one.node_text",
						tfjsonpath.New("result").AtMapKey("data"),
						knownvalue.StringExact("text"),
					),
				},
			},
			{
				Config: testAccQueryOneDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.xpath_query_one.attribute_value",
						tfjsonpath.New("result").AtMapKey("data"),
						knownvalue.StringExact("value"),
					),
				},
			},
			{
				Config: testAccQueryOneDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.xpath_query_one.ns_node_text",
						tfjsonpath.New("result").AtMapKey("data"),
						knownvalue.StringExact("defaultText"),
					),
				},
			},
			{
				Config: testAccQueryOneDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.xpath_query_one.prefix_node_text",
						tfjsonpath.New("result").AtMapKey("data"),
						knownvalue.StringExact("prefixText"),
					),
				},
			},
		},
	})
}

const testAccQueryOneDataSourceConfig = `
locals {
  content = <<EOF
<?xml version="1.0"?>
<root>
	<node attribute="value">text</node>
	<ns xmlns="https://example.com/default" xmlns:prefix="https://example.com/prefix">
		<node>defaultText</node>
		<prefix:node prefix:attribute="prefixValue" attribute="value">prefixText</prefix:node>
	</ns>
</root>
EOF
}

data "xpath_query_one" "node_text" {
  expression = "//node/text()"

  namespace_bindings = {
	"" = "other"
  }

  content = local.content
}

data "xpath_query_one" "attribute_value" {
  expression = "//node/@attribute"

  content = local.content
}

data "xpath_query_one" "ns_node_text" {
  expression = "//node/text()"

  namespace_bindings = {
    ""       = "https://example.com/default"
	"prefix" = "https://example.com/prefix"
  }

  content = local.content
}

data "xpath_query_one" "ns_prefix_node_text" {
  expression = "//prefix:node/text()"

  namespace_bindings = {
    ""       = "https://example.com/default"
	"prefix" = "https://example.com/prefix"
  }

  content = local.content
}
`
