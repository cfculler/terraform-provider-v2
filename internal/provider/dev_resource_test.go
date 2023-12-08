package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDevResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `


resource "devops-bootcamp_dev" "test" {
	name  = "Dev_group"
	engineers = [
		{
			name = "Ryan"
		}
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name
					resource.TestCheckResourceAttr("devops-bootcamp_dev.test", "name", "Dev_group"),
					// Verify engineer values
					resource.TestCheckResourceAttr("devops-bootcamp_dev.test", "engineers.#", "1"),
					resource.TestCheckResourceAttr("devops-bootcamp_dev.test", "engineers.0.name", "Ryan"),
					resource.TestCheckResourceAttr("devops-bootcamp_dev.test", "engineers.0.email", "ryan@ferrets.com"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("devops-bootcamp_dev.test", "id"),
					resource.TestCheckResourceAttrSet("devops-bootcamp_dev.test", "engineers.0.id")),
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "devops-bootcamp_dev" "test" {
	name  = "updatedDev_group"
	engineers = [
		{
			name = "zach"
		}
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name
					resource.TestCheckResourceAttr("devops-bootcamp_dev.test", "name", "updatedDev_group"),
					// Verify engineer values
					resource.TestCheckResourceAttr("devops-bootcamp_dev.test", "engineers.#", "1"),
					resource.TestCheckResourceAttr("devops-bootcamp_dev.test", "engineers.0.name", "zach"),
					resource.TestCheckResourceAttr("devops-bootcamp_dev.test", "engineers.0.email", "zach@bengal.com"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("devops-bootcamp_dev.test", "id"),
					resource.TestCheckResourceAttrSet("devops-bootcamp_dev.test", "engineers.0.id")),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
