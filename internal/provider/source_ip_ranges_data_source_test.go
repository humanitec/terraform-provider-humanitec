package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSourceIPRangesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccSourceIPRangesDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.humanitec_source_ip_ranges.test", "cidr_blocks.0", "34.141.184.227/32"),
					resource.TestCheckResourceAttr("data.humanitec_source_ip_ranges.test", "cidr_blocks.#", "12"),
				),
			},
		},
	})
}

const testAccSourceIPRangesDataSourceConfig = `
data "humanitec_source_ip_ranges" "test" {}
`
