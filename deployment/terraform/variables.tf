terraform {
  experiments = [variable_validation]
}

variable "aws_access_key" {
  type        = string
  description = "AWS access key associated with an IAM user or role"
}

variable "aws_secret_key" {
  type        = string
  description = "The secret key associated with the access key. This is essentially the 'password' for the access key."
}

variable "aws_region" {
  type        = string
  description = "AWS Region to send the request to"
}

variable "name" {
  type        = string
  description = "Name that will be used in resources names and tags"
  default     = "terraform-aws-lambda-proxy-single-instance"
}

variable "create_vpc" {
  type        = bool
  description = "Create personal VPC."
  default     = false
}

variable "vpc_cidr_block" {
  type        = string
  description = "CIDR block for the VPC."
  default     = "10.0.0.0/16"
}

variable "flow_log_enable" {
  type        = bool
  description = "Enable Flow Log for VPC."
  default     = true
}

variable "flow_log_destination" {
  type        = string
  description = "Provides a VPC/Subnet/ENI Flow Log to capture IP traffic for a specific network interface, subnet, or VPC."
  default     = "cloudwatch"

  validation {
    condition     = contains(["cloudwatch", "s3"], var.flow_log_destination)
    error_message = "Logs can be sent only to a CloudWatch Log Group or a S3 Bucket."
  }
}

variable "app_version" {
  type        = string
  description = "AWS Lambda Proxy app version"
  default     = "latest"
}

variable "app_debug" {
  type        = bool
  description = "Enable general debug logging"
  default     = false
}

variable "instance_type" {
  type        = string
  description = "The instance type of the EC2 instance"
  default     = "t3.small"
}

variable "elastic_ip" {
  type        = bool
  description = "Create EIP for instance"
  default     = true
}

variable "lambda_regions" {
  type        = list(string)
  description = "The list of AWS regions names where proxy lambda will be deployed"
  default     = ["ap-northeast-1", "ap-northeast-2", "ap-south-1", "ap-southeast-1", "ap-southeast-2", "ca-central-1", "eu-central-1", "eu-north-1", "eu-west-1", "eu-west-2", "eu-west-3", "sa-east-1", "us-east-1", "us-east-2", "us-west-1", "us-west-2"]
}

variable "lambda_frequency" {
  type        = string
  description = "Frequency to execute Lambda function. If multiple lambda-regions are specified, this will cause traffic to rotate round robin at the interval specified here"
  default     = "5m"
}

variable "lambda_memory" {
  type        = number
  description = "Memory size in MB for Lambda function. Higher memory may allow for faster network throughput"
  default     = 128
}

variable "proxy_debug" {
  type        = bool
  description = "Enable debug logging for proxy"
  default     = false
}

variable "proxy_credentials" {
  type        = string
  description = "Add proxy credentials in format $USERNAME:$PASSWORD"
  default     = null
}

variable "proxy_port" {
  type        = number
  description = "Proxy server port"
  default     = 8080
}

variable "proxy_dns" {
  type        = string
  description = "Specify a DNS server for the proxy server to use for DNS lookups"
  default     = "1.1.1.1"
}

variable "proxy_bypass_domains" {
  type        = list(string)
  description = "Bypass certain domains from using lambda proxy"
  default     = []
}

variable "proxy_cidr_blocks" {
  type        = list(string)
  description = "List of CIDR blocks for proxy access"
  default     = ["0.0.0.0/0"]
}

variable "tunnel_ssh_user" {
  type        = string
  description = "SSH user for tunnel connections from Lambda"
  default     = ""
}

variable "tunnel_ssh_port" {
  type        = number
  description = "SSH port for tunnel connections from Lambda"
  default     = 2222
}

variable "ssh_cidr_blocks" {
  type        = list(string)
  description = "List of CIDR blocks for SSH access"
  default     = []
}
