variable "aws_profile" {
  description = "Name of AWS CLI profile used to create resources under"
}

variable "ssh_pub_key" {
  description = "An SSH public key used to access the EC2 instance running evilginx"
  default     = "~/.ssh/id_rsa.pub"
}

variable "iam_policy_file" {
  description = "IAM policy defining permissions needed to setup and run Lambdas from EC2"
  default     = "../iam/setup.json"
}

variable "package_url" {
  description = "URL of the awslambdaproxy release package"
  default     = "https://github.com/dan-v/awslambdaproxy/releases/download/v0.0.8/awslambdaproxy-linux-x86-64.zip"
}

variable "proxy_regions" {
  description = "AWS regions to launch Lambdas in"
  default     = "us-west-2,us-west-1,us-east-1,us-east-2,ap-southeast-2"
}

variable "instance_type" {
  description = "The instance type of the EC2 instance running awslambdaproxy"
  default     = "t2.micro"
}

variable "ami_id" {
  description = "The AMI to use for the EC2 instance running awslambdaproxy. Defaults to 18.04 LTS 'bionic' HVM EBS store"
  default     = "ami-0532935b53d8e05ee"
}

variable "ec2_user" {
  description = "The username to login as on the EC2 instance running awslambdaproxy"
  default     = "ubuntu"
}
