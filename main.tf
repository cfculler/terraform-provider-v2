terraform {
  required_providers {
    devops-bootcamp = {
      source = "liatr.io/terraform/devops-bootcamp"
    }
  }
}

provider "devops-bootcamp" {}

# data "devops-bootcamp_engineers" "engineers" {}

# output "engineers" {
#   value = devops-bootcamp_engineers.engineers
# }

resource "devops-bootcamp_engineers" "test" {
  name  = "carson"
  email = "carson.culler@liatrio.com"
}

output "engineers" {
  value = devops-bootcamp_engineers.test
}