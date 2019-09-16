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

variable "vpc_cidr" {
  default = "192.168.0.0/16"
}

variable "vswitch_cidr" {
  default = "192.168.0.0/16"
}

variable "vpc_name" {
  default = "vpc_ecs_userdata"
}

variable "availability_zone" {
  default = "cn-beijing-f"
}

variable "sg_name" {
  default = "sg_ecs_userdata"
}

variable "rule_policy" {
  default = "accept"
}

variable "image" {
  default = "ubuntu_18_04_64_20G_alibase_20190624.vhd"
}

variable "ecs_type" {
  default = "ecs.t5-lc2m1.nano"
}

variable "password" {}

# Output
output "ecs_id" {
  value = "${alicloud_instance.nginx.id}"
}

output "ecs_public_ip" {
  value = "${alicloud_instance.nginx.public_ip}"
}

# Resources
resource "alicloud_vpc" "main" {
  cidr_block = "${var.vpc_cidr}"
  name = "${var.vpc_name}"
}

resource "alicloud_vswitch" "main" {
  vpc_id = "${alicloud_vpc.main.id}"
  availability_zone = "${var.availability_zone}"
  cidr_block = "${var.vswitch_cidr}"

  depends_on = [
    "alicloud_vpc.main"
  ]
}

resource "alicloud_security_group" "main" {
  name = "${var.sg_name}"
  description = "Default security group for VPC"
  vpc_id = "${alicloud_vpc.main.id}"
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

resource "alicloud_instance" "nginx" {
  image_id = "${var.image}"
  vswitch_id = "${alicloud_vswitch.main.id}"
  availability_zone = "${var.availability_zone}"

  instance_type = "${var.ecs_type}"
  system_disk_category = "cloud_efficiency"
  system_disk_size = 20

  internet_charge_type = "PayByTraffic"
  internet_max_bandwidth_out = 5
  security_groups = [
    "${alicloud_security_group.main.id}",
  ]
  instance_name = "nginx"
  password = "${var.password}"

  user_data = "${file("launch.sh")}"
}
