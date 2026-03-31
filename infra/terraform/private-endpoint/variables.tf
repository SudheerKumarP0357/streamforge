variable "resource_group_name" {
  type = string
}

variable "location" {
  type = string
}

variable "subnet_id" {
  type = string
}

variable "private_endpoint_name" {
  type = string
}

variable "resource_id" {
  type = string
}

variable "subresource_names" {
  type = list(string)
}

variable "tags" {
  type = object({
    Environment = string
    Owner       = string
    Created_By  = string
    Application = string
  })
}

variable "private_dns_zone_name" {
  type = string
}

variable "jump_vnet_id" {
  type = string
}

variable "app_vnet_id" {
  type = string
}
