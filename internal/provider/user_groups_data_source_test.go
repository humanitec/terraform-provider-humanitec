package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var (
	idpId   = os.Getenv("HUMANITEC_TERRAFORM_ORG_IDP_ID")
	id      = os.Getenv("HUMANITEC_TERRAFORM_ORG_GROUP_ID")
	groupId = os.Getenv("HUMANITEC_TERRAFORM_ORG_GROUP_GROUP_ID")
)

func TestAccUserGroupsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccCreateUserGroupsDataSourceConfigFilterById(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.humanitec_user_groups.test", "groups.0.id"),
				),
			},
			{
				Config: testAccCreateUserGroupsDataSourceConfigFilterById(id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.humanitec_user_groups.test", "groups.0.id"),
				),
			},
			{
				Config: testAccCreateUserGroupsDataSourceConfigFilterByIdp(idpId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.humanitec_user_groups.test", "groups.0.id"),
				),
			},
			{
				Config: testAccCreateUserGroupsDataSourceConfigFilterByIdpAndGroupId(idpId, groupId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.humanitec_user_groups.test", "groups.0.id"),
				),
			},
			{
				Config: testAccCreateUserGroupsDataSourceConfigFilterByIdp("not-existing-idp-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("data.humanitec_user_groups.test", "groups.0"),
				),
			},
			{
				Config: testAccCreateUserGroupsDataSourceConfigFilterByIdpAndGroupId(idpId, "not-existing-group-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("data.humanitec_user_groups.test", "groups.0"),
				),
			},
		},
	})
}

func testAccCreateUserGroupsDataSourceConfigFilterById(id string) string {
	filtersString := ""
	if id != "" {
		filtersString = fmt.Sprintf(`filter = {
			id = "%s"
		}`, id)
	}
	return fmt.Sprintf(`data "humanitec_user_groups" "test" {
		%s
	}`, filtersString)
}

func testAccCreateUserGroupsDataSourceConfigFilterByIdp(idpId string) string {
	return fmt.Sprintf(`data "humanitec_user_groups" "test" {
		filter = {
			idp_id = "%s"
		}
	}`, idpId)
}

func testAccCreateUserGroupsDataSourceConfigFilterByIdpAndGroupId(idpId, groupId string) string {
	return fmt.Sprintf(`data "humanitec_user_groups" "test" {
		filter = {
			idp_id = "%s"
			group_id = "%s"
		}
	}`, idpId, groupId)
}
