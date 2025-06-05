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
  chosen_image    = try([for image in data.arvan_images.terraform_image.distributions : image if image.distro_name == var.chosen_distro_name && image.name == var.chosen_name][0], null)
  selected_plan_k8s = try([for plan in data.arvan_plans.plan_list.plans : plan if plan.id == var.chosen_plan_id_k8s][0], null)
  selected_plan_ha  = try([for plan in data.arvan_plans.plan_list.plans : plan if plan.id == var.chosen_plan_id_ha][0], null)
}

data "arvan_networks" "terraform_network" {
  region = var.region
}

locals {
  network_list = tolist(data.arvan_networks.terraform_network.networks)
  chosen_network = try(
    [for network in local.network_list : network
    if network.name == var.chosen_network_name],
    []
  )
}

output "chosen_network" {
  value = local.chosen_network
}

resource "arvan_network" "terraform_private_network" {
  region      = var.region
  name        = "tf_private_network"
  dhcp_range = {
    start = "10.255.255.100"
    end   = "10.255.255.200"
  }
  dns_servers    = ["8.8.8.8", "1.1.1.1"]
  enable_dhcp    = true
  enable_gateway = true
  cidr           = "10.255.255.0/24"
  gateway_ip     = "10.255.255.1"
}

resource "arvan_abrak" "masters" {
  depends_on = [arvan_network.terraform_private_network]
  count        = 3
  region       = var.region
  name         = "master_${count.index + 1}"
  image_id     = local.chosen_image.id
  flavor_id    = local.selected_plan_k8s.id
  ssh_key_name = "macbook"
  disk_size    = 25
  networks = [
    {
      network_id = arvan_network.terraform_private_network.network_id
    }
  ]

  security_groups = [data.arvan_security_groups.default_security_groups.groups[0].id]
  timeouts {
    create = "1h30m"
    update = "2h"
    delete = "20m"
    read   = "10m"
  }
}

resource "arvan_abrak" "haproxy" {
  depends_on = [arvan_network.terraform_private_network]
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
  networks = [
    {
      network_id = local.chosen_network[0].network_id
    },
    {
      network_id = arvan_network.terraform_private_network.network_id
    }
  ]
  security_groups = [data.arvan_security_groups.default_security_groups.groups[0].id]
}

