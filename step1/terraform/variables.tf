variable "api_key" {
  type      = string
  sensitive = true
}

variable "region" {
  type    = string
  default = "ir-thr-ba1"
}

variable "chosen_distro_name" {
  type    = string
  default = "ubuntu"
}

variable "chosen_name" {
  type    = string
  default = "24.04"
}

variable "chosen_plan_id" {
  type    = string
  default = "g2-12-4-0"
}

