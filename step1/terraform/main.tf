terraform {
  required_providers {
    arvan = {
      source = "terraform.arvancloud.ir/arvancloud/iaas"
    }
  }
}

provider "arvan" {
  api_key = var.api_key
}

data "arvan_security_groups" "default_security_groups" {
  region = var.region
}

data "arvan_images" "terraform_image" {
  region     = var.region
  image_type = "distributions" // or one of: arvan, private
}

data "arvan_plans" "plan_list" {
  region = var.region
}

locals {
  chosen_image = try([for image in data.arvan_images.terraform_image.distributions : image
  if image.distro_name == var.chosen_distro_name && image.name == var.chosen_name][0], null)
  selected_plan = try([for plan in data.arvan_plans.plan_list.plans : plan if plan.id == var.chosen_plan_id][0], null)
}

data "arvan_networks" "terraform_network" {
  region = var.region
}

resource "arvan_abrak" "abrak" {
  timeouts {
    create = "1h30m"
    update = "2h"
    delete = "20m"
    read   = "10m"
  }
  region       = var.region
  name         = "abrak-${count.index + 1}"
  ssh_key_name = "macbook"
  count        = 2
  image_id     = local.chosen_image.id
  flavor_id    = local.selected_plan.id
  disk_size    = 25
  security_groups = [data.arvan_security_groups.default_security_groups.groups[0].id]
}
