// package provider

// import (
// 	"testing"

// 	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
// )

// func TestAccEngineerResource(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			// Create and Read testing
// 			{
// 				Config: `provider "devops-bootcamp" {}
// 				` + `
// resource "devops-bootcamp_engineers" "test" {
//   name = test
//   email = test@test.com
// }
// `,
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr("devops-bootcamp_engineers.test", "name", "test"),
// 					resource.TestCheckResourceAttr("devops-bootcamp_engineers.test", "email", "test@test.com"),
// 					resource.TestCheckResourceAttrSet("devops-bootcamp_engineers.test", "id"),
// 					resource.TestCheckResourceAttrSet("devops-bootcamp_engineers.test", "last_updated"),
// 				),
// 			},
// 			// ImportState testing
// 			{
// 				ResourceName:      "devops-bootcamp_engineers.test",
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 				// The last_updated attribute does not exist in the HashiCups
// 				// API, therefore there is no value for it during import.
// 				ImportStateVerifyIgnore: []string{"last_updated"},
// 			},
// 			// Update and Read testing
// 			{
// 				Config: `provider "devops-bootcamp" {}
// 				` + `
// resource "devops-bootcamp_engineers" "test" {
// 	name = test-changed
// 	email = test@test.com
// }
// `,
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr("devops-bootcamp_engineers.test", "name", "test-changed"),
// 					resource.TestCheckResourceAttr("devops-bootcamp_engineers.test", "email", "test@test.com"),
// 					resource.TestCheckResourceAttrSet("devops-bootcamp_engineers.test", "id"),
// 					resource.TestCheckResourceAttrSet("devops-bootcamp_engineers.test", "last_updated"),
// 				),
// 			},
// 			// Delete testing automatically occurs in TestCase
// 		},
// 	})
// }

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestEngineerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
                    provider "devops-bootcamp" {}

                    resource "devops-bootcamp_engineer" "test" {
                        name  = "Bobby"
                        email = "Bobby@gmail.com"
                    }
		`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name
					resource.TestCheckResourceAttr("devops-bootcamp_engineer.test", "name", "Bobby"),
					// Verify email
					resource.TestCheckResourceAttr("devops-bootcamp_engineer.test", "email", "Bobby@gmail.com"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("devops-bootcamp_engineer.test", "id")),
			},
			// Update and Read testing
			{
				Config: `

                provider "devops-bootcamp" {}

                resource "devops-bootcamp_engineer" "test" {
                    name  = "updatedBobby"
                    email = "updatedBobby@gmail.com"
                }
	`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name/email updated
					resource.TestCheckResourceAttr("devops-bootcamp_engineer.test", "name", "updatedBobby"),
					resource.TestCheckResourceAttr("devops-bootcamp_engineer.test", "email", "updatedBobby@gmail.com"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
