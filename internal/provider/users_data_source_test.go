package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUsersDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccCreateUsersDataSourceConfig(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.humanitec_users.test", "users.0.id"),
				),
			},
			{
				Config: testAccCreateUsersDataSourceConfig("not-exsting-email-address@humanitec.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("data.humanitec_users.test", "users.0"),
				),
			},
		},
	})
}

func testAccCreateUsersDataSourceConfig(email string) string {
	filtersString := ""
	if email != "" {
		filtersString = fmt.Sprintf(`filter = {
			email = "%s"
		}`, email)
	}
	return fmt.Sprintf(`data "humanitec_users" "test" {
		%s
	}`, filtersString)
}
