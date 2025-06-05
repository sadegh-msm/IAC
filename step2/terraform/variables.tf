variable "api_key" {
  type      = string
  sensitive = true
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

variable "chosen_network_name" {
  type        = string
  description = "The chosen name of network"
  default     = "public210" //public202
}

variable "chosen_plan_id_k8s" {
  type        = string
  default     = "g2-8-4-0"
}

variable "chosen_plan_id_ha" {
  type        = string
  default     = "g6-2-2-0"
}

