package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceServiceUserToken(t *testing.T) {
	const (
		userName    = "test user"
		tokenId     = "test-token-id"
		description = "Test token description"
	)
	expiresAt := time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04:05.999Z")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCreateResourceServiceUserToken(userName, tokenId, "", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_service_user_token.token", "id", tokenId),
					resource.TestCheckResourceAttrSet("humanitec_service_user_token.token", "token"),
					resource.TestCheckNoResourceAttr("humanitec_service_user_token.token", "description"),
					resource.TestCheckNoResourceAttr("humanitec_service_user_token.token", "expires_at"),
				),
			},
			// Update and Read testing
			{
				Config: testAccCreateResourceServiceUserToken(userName, tokenId, description, expiresAt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_service_user_token.token", "id", tokenId),
					resource.TestCheckResourceAttrSet("humanitec_service_user_token.token", "token"),
					resource.TestCheckResourceAttr("humanitec_service_user_token.token", "description", description),
					resource.TestCheckResourceAttr("humanitec_service_user_token.token", "expires_at", expiresAt),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccCreateResourceServiceUserToken(userName, tokenId, tokenDescription, tokenExpiration string) string {
	tokenDescriptionEntry := ""
	if tokenDescription != "" {
		tokenDescriptionEntry = fmt.Sprintf(`description = "%s"`, tokenDescription)
	}
	tokenExpirationEntry := ""
	if tokenExpiration != "" {
		tokenExpirationEntry = fmt.Sprintf(`expires_at = "%s"`, tokenExpiration)
	}

	return fmt.Sprintf(`
	resource "humanitec_user" "service_user" {
		name = "%s"
		role = "administrator"
		type = "service"
	  }
	  
	  resource "humanitec_service_user_token" "token" {
		id = "%s"
		user_id = humanitec_user.service_user.id
		%s
		%s
	  }`,
		userName, tokenId, tokenDescriptionEntry, tokenExpirationEntry,
	)
}
