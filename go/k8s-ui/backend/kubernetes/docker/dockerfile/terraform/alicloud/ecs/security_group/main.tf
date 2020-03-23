
# Provider
provider "alicloud" {
  version = "~> 1.48"
  access_key = "${var.AccessKeyID}"
  secret_key = "${var.AccessKeySecret}"
}


# Variables
variable "AccessKeyID" {}

variable "AccessKeySecret" {}

variable "sg_name" {
  default = "alicloud_sg_1"
}

variable "vpc_id" {}

variable "rule_policy" {
  default = "accept"
}

# Resources
resource "alicloud_security_group" "main" {
  name = "${var.sg_name}"
  description = "Default security group for VPC"
  vpc_id = "${var.vpc_id}"
}

resource "alicloud_security_group_rule" "ssh-in" {
  nic_type = "intranet"
  type = "ingress"
  policy = "${var.rule_policy}"
  ip_protocol = "tcp"
  port_range = "22/22"
  priority = 1
  cidr_ip = "0.0.0.0/0"
  security_group_id = "${alicloud_security_group.main.id}"
}

resource "alicloud_security_group_rule" "http-in" {
  nic_type = "intranet"
  type = "ingress"
  policy = "${var.rule_policy}"
  ip_protocol = "tcp"
  port_range = "80/80"
  priority = 1
  cidr_ip = "0.0.0.0/0"
  security_group_id = "${alicloud_security_group.main.id}"
}

resource "alicloud_security_group_rule" "https-in" {
  nic_type = "intranet"
  type = "ingress"
  policy = "${var.rule_policy}"
  ip_protocol = "tcp"
  port_range = "443/443"
  priority = 1
  cidr_ip = "0.0.0.0/0"
  security_group_id = "${alicloud_security_group.main.id}"
}
