terraform {
  required_providers {
    devops-bootcamp = {
      source = "liatr.io/terraform/devops-bootcamp"
    }
  }
}

provider "devops-bootcamp" {}

resource "devops-bootcamp_engineers" "engineers" {
  name  = "carson"
  email = "carsonculler@liatrio.com"
}

# output "engineers" {
#   value = data.devops-bootcamp_engineer.engineers
# }