variable "aws_profile" {}
variable "ssh-pub-key" {
  default = "~/.ssh/id_rsa.pub"
}

#ssh-pub-key = "id_rsa.pub"

provider "aws" {
  profile = var.aws_profile
  region  = "ap-southeast-2"
}

resource "aws_key_pair" "lambdaproxy_ssh_key" {
  key_name   = "lambdaproxy_ssh_key"
  public_key = "${file("${var.ssh-pub-key}")}"
}

resource "aws_vpc" "lambdaproxy" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
}

resource "aws_subnet" "lambdaproxy" {
  vpc_id     = "${aws_vpc.lambdaproxy.id}"
  cidr_block = "10.0.0.0/24"
}

resource "aws_internet_gateway" "lambdaproxy" {
  vpc_id = "${aws_vpc.lambdaproxy.id}"
}

resource "aws_route_table" "lambdaproxy" {
  vpc_id = "${aws_vpc.lambdaproxy.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.lambdaproxy.id}"
  }
}

resource "aws_route_table_association" "lambdaproxy" {
  subnet_id      = "${aws_subnet.lambdaproxy.id}"
  route_table_id = "${aws_route_table.lambdaproxy.id}"
}

resource "aws_security_group" "lambdaproxy" {
  name   = "lambdaproxy-ports"
  vpc_id = "${aws_vpc.lambdaproxy.id}"

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

}

resource "aws_instance" "lambdaproxybox" {
  # ap-southeast-2 18.04 LTS "bionic" HVM EBS store
  ami           = "ami-0532935b53d8e05ee"
  instance_type = "t2.micro"
  key_name      = "${aws_key_pair.lambdaproxy_ssh_key.key_name}"
  vpc_security_group_ids = [
    "${aws_security_group.lambdaproxy.id}"
  ]
  subnet_id                   = "${aws_subnet.lambdaproxy.id}"
  associate_public_ip_address = true
}

output "lambdaproxybox-ec2-ip" {
  value = "${aws_instance.lambdaproxy.public_ip}"
}
