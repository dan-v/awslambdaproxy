provider "aws" {
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = var.aws_region
}

provider "aws" {
  alias = "ap-northeast-1"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "ap-northeast-1"
}

provider "aws" {
  alias = "ap-northeast-2"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "ap-northeast-2"
}

provider "aws" {
  alias = "ap-south-1"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "ap-south-1"
}

provider "aws" {
  alias = "ap-southeast-1"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "ap-southeast-1"
}

provider "aws" {
  alias = "ap-southeast-2"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "ap-southeast-2"
}

provider "aws" {
  alias = "ca-central-1"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "ca-central-1"
}

provider "aws" {
  alias = "eu-central-1"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "eu-central-1"
}

provider "aws" {
  alias = "eu-north-1"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "eu-north-1"
}

provider "aws" {
  alias = "eu-west-1"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "eu-west-1"
}

provider "aws" {
  alias = "eu-west-2"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "eu-west-2"
}

provider "aws" {
  alias = "eu-west-3"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "eu-west-3"
}

provider "aws" {
  alias = "sa-east-1"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "sa-east-1"
}

provider "aws" {
  alias = "us-east-1"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "us-east-1"
}

provider "aws" {
  alias = "us-east-2"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "us-east-2"
}

provider "aws" {
  alias = "us-west-1"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "us-west-1"
}

provider "aws" {
  alias = "us-west-2"

  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = "us-west-2"
}
