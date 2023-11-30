terraform {
  required_providers {
    devops-bootcamp = {
      source = "liatr.io/terraform/devops-bootcamp"
    }
  }
}

provider "devops-bootcamp" {}

data "devops-bootcamp_engineers" "engineers" {}

# output "engineers" {
#   value = data.devops-bootcamp_engineer.engineers
# }