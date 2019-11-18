provider "aws" {
  access_key = "${var.AccessKeyID_lx20081036}"
  secret_key = "${var.AccessKeySecret_lx20081036}"
  region = "${var.region}"
}

variable "AccessKeyID" {}

variable "AccessKeySecret" {}

variable "AccessKeyID_lx20081036" {} // free account

variable "AccessKeySecret_lx20081036" {}

variable "region" {
  default = "ap-northeast-1"
}

variable "cidr_block_range" {
  description = "The CIDR block for the VPC"
  default = "10.1.0.0/16"
}

variable "environment_tag" {
  description = "Environment tag"
  default = "terraform"
}

variable "subnet_cidr_block" {
  description = "The CIDR block for public subnet of VPC"
  default = "10.1.0.0/24"
}

variable "availability_zone" {
  default = "ap-northeast-1a"
}

variable "public_key_path" {
  description = "Public key path"
  default = "~/.ssh/id_rsa.pub"
}

variable "vpc_name" {
  default = "packer-ansible-terraform"
}

variable "aws_internet_gateway_name" {
  default = "packer-ansible-terraform"
}

variable "aws_subnet_name" {
  default = "packer-ansible-terraform"
}

variable "aws_route_table_name" {
  default = "packer-ansible-terraform"
}


resource "aws_vpc" "main" {
  cidr_block = "${var.cidr_block_range}"
  enable_dns_support = true
  enable_dns_hostnames = true

  tags {
    Name = "${var.vpc_name}"
    Environment = "${var.environment_tag}"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id = "${aws_vpc.main.id}"

  tags {
    Name = "${var.aws_internet_gateway_name}"
    Environment = "${var.environment_tag}"
  }
}

resource "aws_subnet" "main" {
  cidr_block = "${var.subnet_cidr_block}"
  vpc_id = "${aws_vpc.main.id}"
  map_public_ip_on_launch = true
  availability_zone = "${var.availability_zone}"

  tags {
    Name = "${var.aws_subnet_name}"
    Environment = "${var.environment_tag}"
    Type = "Public"
  }
}

resource "aws_route_table" "main" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.main.id}"
  }

  tags {
    Name = "${var.aws_route_table_name}"
    Environment = "${var.environment_tag}"
  }
}

resource "aws_route_table_association" "main" {
  route_table_id = "${aws_route_table.main.id}"
  subnet_id = "${aws_subnet.main.id}"
}

resource "aws_key_pair" "main" {
  key_name = "public key"
  public_key = "${file(var.public_key_path)}"
}

resource "aws_key_pair" "lx1036@github.com" {
  key_name = "lx1036@github.com"
  public_key = "${file(var.public_key_path)}"
}

data "terraform_remote_state" "network" {
  backend = "local"

  config {
    path="./terraform.tfstate"
  }
}

data "aws_ami" "ec2-ami" {
  owners = ["self"]

  filter {
    name = "state"
    values = ["available"]
  }

  filter {
    name = "tag:Name"
    values = ["lx1036-Packer-Ansible-AMI"]
  }

  most_recent = true
}

module "securityGroupModule" {
  source = "./terraform/securityGroup"
  access_key = "${var.AccessKeyID_lx20081036}"
  secret_key = "${var.AccessKeySecret_lx20081036}"
  region = "${var.region}"
  vpc_id = "${data.terraform_remote_state.network.vpc_id}"
  environment_tag = "${var.environment_tag}"
}

module "instanceModule" {
  source = "./terraform/instance"
  access_key = "${var.AccessKeyID_lx20081036}"
  secret_key = "${var.AccessKeySecret_lx20081036}"
  region = "${var.region}"

  subnet_public_id = "${data.terraform_remote_state.network.public_subnets[0]}"
  key_name = "${data.terraform_remote_state.network.ec2keyName}"
  security_group_ids = ["${module.securityGroupModule.sg_22}", "${module.securityGroupModule.sg_80}"]

  instance_ami = "${data.aws_ami.ec2-ami.id}" // built by Packer and Ansible

  environment_tag = "${var.environment_tag}"
}

output "vpc_id" {
  value = "${aws_vpc.main.id}"
}
output "public_subnets" {
  value = ["${aws_subnet.main.id}"]
}
output "ec2keyName" {
  value = "${aws_key_pair.main.key_name}"
}

output "instance_id" {
  value = "${module.instanceModule.instance}"
}

output "eip_public" {
  value = "${module.instanceModule.eip_public}"
}

output "eip_private" {
  value = "${module.instanceModule.eip_private}"
}
