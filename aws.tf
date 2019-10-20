variable "aws-profile" {}
variable "ssh-pub-key" {
  default = "~/.ssh/id_rsa.pub"
}
variable "iam-policy-file" {
  default = "./iam/setup.json"
}
variable "package-url" {
  default = "https://github.com/dan-v/awslambdaproxy/releases/download/v0.0.8/awslambdaproxy-linux-x86-64.zip"
}
variable "proxy-regions" {
  default = "us-west-2,us-west-1,us-east-1,us-east-2,ap-southeast-2"
}

provider "aws" {
  profile = var.aws-profile
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
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_iam_role" "assume_role" {
  name = "test_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "lambdaproxy-instance-profile" {
  name = "lambdaproxy-instance-profile"
  role = "${aws_iam_role.assume_role.name}"
}

resource "aws_iam_role_policy" "setup_policy" {
  name = "setup_policy"
  role = "${aws_iam_role.assume_role.id}"

  # This policy's privileges cover both the setup and runtime cases
  policy = "${file("${var.iam-policy-file}")}"
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
  iam_instance_profile        = "${aws_iam_instance_profile.lambdaproxy-instance-profile.name}"

  tags = {
    Name = "lambdaproxy"
  }
}

# When you specify the remote-exec within the aws_instance block
# Terraform will run that code before the security group is attached
# which is a completely braindead idea because, y'know, ya might need
# freakin Internet access when you're provisioning your instance.
resource "null_resource" "foobar" {
  triggers = {
    public_ip = "${aws_instance.lambdaproxybox.public_ip}"
  }

  # Terraform's ridiculous default approach to ssh'ing into the
  # instance with the root account doesn't play nicely with our
  # Ubuntu AMI, so this:
  connection {
    type  = "ssh"
    host  = "${aws_instance.lambdaproxybox.public_ip}"
    user  = "ubuntu"
    port  = "22"
    agent = true
  }

  provisioner "remote-exec" {
    inline = [
      "wget ${var.package-url}",
      "sudo apt install -y unzip",
      "unzip awslambdaproxy-linux-x86-64.zip",
      "./awslambdaproxy setup"
    ]
  }
}

output "lambdaproxybox-ec2-ip" {
  value = "${aws_instance.lambdaproxybox.public_ip}"
}
