terraform {
  required_providers {
    arvan = {
      source = "terraform.arvancloud.ir/arvancloud/iaas"
    }
  }
}

variable "ApiKey" {
  type      = string
  default   = "test"
  sensitive = true
}

provider "arvan" {
  api_key = var.ApiKey
}

variable "region" {
  type        = string
  default     = "ir-thr-ba1"
}

variable "chosen_distro_name" {
  type        = string
  default     = "ubuntu"
}

variable "chosen_name" {
  type        = string
  default     = "24.04"
}

# variable "chosen_plan_id" {
#   type        = string
#   default     = "g2-4-2-0"
# }

variable "chosen_plan_id_k8s" {
  type        = string
  default     = "g2-12-4-0"
}

variable "chosen_plan_id_ha" {
  type        = string
  default     = "g2-4-2-0"
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
  chosen_image    = try([for image in data.arvan_images.terraform_image.distributions : image if image.distro_name == var.chosen_distro_name && image.name == var.chosen_name][0], null)
  selected_plan_k8s = try([for plan in data.arvan_plans.plan_list.plans : plan if plan.id == var.chosen_plan_id_k8s][0], null)
  selected_plan_ha  = try([for plan in data.arvan_plans.plan_list.plans : plan if plan.id == var.chosen_plan_id_ha][0], null)
}

data "arvan_networks" "terraform_network" {
  region = var.region
}

resource "arvan_abrak" "masters" {
  count        = 3
  region       = var.region
  name         = "master_${count.index + 1}"
  image_id     = local.chosen_image.id
  flavor_id    = local.selected_plan_k8s.id
  ssh_key_name = "macbook"
  disk_size    = 25
  security_groups = [data.arvan_security_groups.default_security_groups.groups[0].id]
  timeouts {
    create = "1h30m"
    update = "2h"
    delete = "20m"
    read   = "10m"
  }
}

resource "arvan_abrak" "haproxy" {
  timeouts {
    create = "1h30m"
    update = "2h"
    delete = "20m"
    read   = "10m"
  }
  region       = var.region
  name         = "haproxy"
  ssh_key_name = "macbook"
  count        = 1
  image_id     = local.chosen_image.id
  flavor_id    = local.selected_plan_ha.id
  disk_size    = 25
  security_groups = [data.arvan_security_groups.default_security_groups.groups[0].id]
}

