provider "alicloud" {
  version = "1.46.0"
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
  region = "cn-beijing"
}

data "alicloud_instance_types" "c1g1" {
  cpu_core_count = 1
  memory_size = 1
}

data "alicloud_images" "default" {
  name_regex = "^ubuntu"
  most_recent = true
  owners = "system"
}

# Create security group
resource "alicloud_security_group" "default" {
  name = "test2"
  description = "test2"
//  vpc_id = "vpc-abc123"
}
