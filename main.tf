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

resource "devops-bootcamp_engineers" "engineers" {
  name  = "carson-two"
  email = "carsonculler@liatrio.com"
}

# resource "devops-bootcamp_engineers" "test" {
#   name  = "carson"
#   email = "test@liatrio.com"
# }
