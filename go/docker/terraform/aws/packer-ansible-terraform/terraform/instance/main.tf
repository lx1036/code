provider "aws" {
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
  region = "${var.region}"
}

variable "access_key" {}

variable "secret_key" {}

variable "region" {}

variable "instance_ami" {
  default = "ami-09c81ecf1c2b2ef70" // Ubuntu Server 18.04 LTS (HVM), SSD Volume Type
  description = "EC2 instance ami"
}

variable "instance_type" {
  default = "t2.micro" // available for free
  description = "EC2 instance type"
}

variable "subnet_public_id" {
  description = "VPC public subnet id"
  default = ""
}

variable "key_name" {}

variable "environment_tag" {}

variable "security_group_ids" {
  description = "EC2 ssh/http security group"
  type = "list"
  default = []
}

resource "aws_instance" "instance" {
  ami = "${var.instance_ami}"
  instance_type = "${var.instance_type}"
  subnet_id = "${var.subnet_public_id}"
  vpc_security_group_ids = ["${var.security_group_ids}"]
  key_name = "${var.key_name}"

  tags {
    Name = "terraform"
    Environment = "${var.environment_tag}"
  }
}

resource "aws_eip" "instance_eip" {
  vpc = true
  instance = "${aws_instance.instance.id}"

  tags {
    Environment = "${var.environment_tag}"
  }
}

output "instance" {
  value = "${aws_instance.instance.id}"
}

output "eip_public" {
  value = "${aws_eip.instance_eip.public_ip}"
}

output "eip_private" {
  value = "${aws_eip.instance_eip.private_ip}"
}
