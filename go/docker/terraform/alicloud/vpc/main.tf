# tf apply/plan

# Provider
provider "alicloud" {
  version = "~> 1.48"
  access_key = "${var.AccessKeyID}"
  secret_key = "${var.AccessKeySecret}"
  region = "cn-beijing"
}

# Variables
variable "AccessKeyID" {}

variable "AccessKeySecret" {}

variable "vpc_name" {
  default = "alicloud_vpc_1"
}

variable "cidr_blocks" {
  type = "map"

  default = {
    az0="192.168.1.0/25"
    az1="192.168.2.0/24"
    az2="192.168.4.0/23"
  }
}

variable "vpc_cidr" {
  default = "192.168.0.0/16"
}

variable "availability_zone" {
  default = "cn-beijing-c"
}


# Output
output "vpc_id" {
  value = "${alicloud_vpc.main.id}"
}

output "vswitch_ids" {
  value = "${join(",", alicloud_vswitch.main.*.id)}"
}

output "availability_zones" {
  value = "${join(",", alicloud_vswitch.main.*.availability_zone)}"
}


# Resources
resource "alicloud_vpc" "main" {
  cidr_block = "${var.vpc_cidr}"
  name = "${var.vpc_name}"
}

resource "alicloud_vswitch" "main" {
  vpc_id = "${alicloud_vpc.main.id}"
  availability_zone = "${var.availability_zone}"
  count="${length(var.cidr_blocks)}"
  cidr_block = "${lookup(var.cidr_blocks, "az${count.index}")}"

  depends_on = [
    "alicloud_vpc.main"
  ]
}

/*resource "alicloud_nat_gateway" "main" {
  vpc_id = "${alicloud_vpc.main.id}"
  specification = "small"
  name = "nat-from-terraform"
}

resource "alicloud_eip" "main" {

}

resource "alicloud_eip_association" "main" {
  allocation_id = "${alicloud_eip.main.id}"
  instance_id = "${alicloud_nat_gateway.main.id}"
}*/
