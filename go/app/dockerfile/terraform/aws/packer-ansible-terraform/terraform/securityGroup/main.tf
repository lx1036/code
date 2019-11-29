provider "aws" {
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
  region = "${var.region}"
}

variable "access_key" {}

variable "secret_key" {}

variable "region" {}

variable "vpc_id" {
  default = ""
}

variable "environment_tag" {
  default = ""
}

resource "aws_security_group" "sg_22" {
  name = "sg_22"
  description = "SSH"
  vpc_id = "${var.vpc_id}"

  ingress {
    from_port = 22
    protocol = "tcp"
    to_port = 22
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port = 0
    protocol = "-1"
    to_port = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Environment = "${var.environment_tag}"
  }
}

resource "aws_security_group" "sg_80" {
  name = "sg_80"
  description = "HTTP"
  vpc_id = "${var.vpc_id}"

  ingress {
    from_port = 80
    protocol = "tcp"
    to_port = 80
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port = 0
    protocol = "-1"
    to_port = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Environment = "${var.environment_tag}"
  }
}

output "sg_22" {
  value = "${aws_security_group.sg_22.id}"
}

output "sg_80" {
  value = "${aws_security_group.sg_80.id}"
}
