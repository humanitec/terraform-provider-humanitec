package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var registeredFakeIdpId = os.Getenv("HUMANITEC_TERRAFORM_ORG_IDP_ID")

func TestAccResourceUserGroupUpdateRole(t *testing.T) {
	const (
		groupId = "org-administrators"
		role    = "administrator"
		newRole = "manager"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCreateResourceUserGroup(registeredFakeIdpId, groupId, role),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_user_group.test", "idp_id", registeredFakeIdpId),
					resource.TestCheckResourceAttr("humanitec_user_group.test", "group_id", groupId),
					resource.TestCheckResourceAttr("humanitec_user_group.test", "role", role),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_user_group.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					id := s.RootModule().Resources["humanitec_user_group.test"].Primary.Attributes["id"]
					return id, nil
				},
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccCreateResourceUserGroup(registeredFakeIdpId, groupId, newRole),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_user_group.test", "idp_id", registeredFakeIdpId),
					resource.TestCheckResourceAttr("humanitec_user_group.test", "group_id", groupId),
					resource.TestCheckResourceAttr("humanitec_user_group.test", "role", newRole),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceUserGroupUpdateGroupId(t *testing.T) {
	const (
		groupId    = "org-managers"
		role       = "manager"
		newGroupId = "managers"
	)
	var groupComputedId string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCreateResourceUserGroup(registeredFakeIdpId, groupId, role),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_user_group.test", "idp_id", registeredFakeIdpId),
					resource.TestCheckResourceAttr("humanitec_user_group.test", "group_id", groupId),
					resource.TestCheckResourceAttr("humanitec_user_group.test", "role", role),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_user_group.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					groupComputedId = s.RootModule().Resources["humanitec_user_group.test"].Primary.Attributes["id"]
					return groupComputedId, nil
				},
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccCreateResourceUserGroup(registeredFakeIdpId, newGroupId, role),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_user_group.test", "idp_id", registeredFakeIdpId),
					resource.TestCheckResourceAttr("humanitec_user_group.test", "group_id", newGroupId),
					resource.TestCheckResourceAttr("humanitec_user_group.test", "role", role),
				),
			},
			// ImportState with the previous id should not work anymore
			{
				ResourceName: "humanitec_user_group.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return groupComputedId, nil
				},
				ExpectError: regexp.MustCompile("unexpected status code: 403"),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccCreateResourceUserGroup(idpId, groupId, role string) string {
	return fmt.Sprintf(`
resource "humanitec_user_group" "test" {
	idp_id = "%s"
	group_id = "%s"
	role = "%s"
}
`, idpId, groupId, role)
}
