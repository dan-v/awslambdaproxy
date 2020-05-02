resource "aws_lambda_function" "ap-northeast-1" {
  count    = contains(var.lambda_regions, "ap-northeast-1") ? 1 : 0
  provider = aws.ap-northeast-1

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "ap-northeast-2" {
  count    = contains(var.lambda_regions, "ap-northeast-2") ? 1 : 0
  provider = aws.ap-northeast-2

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "ap-south-1" {
  count    = contains(var.lambda_regions, "ap-south-1") ? 1 : 0
  provider = aws.ap-south-1

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "ap-southeast-1" {
  count    = contains(var.lambda_regions, "ap-southeast-1") ? 1 : 0
  provider = aws.ap-southeast-1

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "ap-southeast-2" {
  count    = contains(var.lambda_regions, "ap-southeast-2") ? 1 : 0
  provider = aws.ap-southeast-2

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "ca-central-1" {
  count    = contains(var.lambda_regions, "ca-central-1") ? 1 : 0
  provider = aws.ca-central-1

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "eu-central-1" {
  count    = contains(var.lambda_regions, "eu-central-1") ? 1 : 0
  provider = aws.eu-central-1

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "eu-north-1" {
  count    = contains(var.lambda_regions, "eu-north-1") ? 1 : 0
  provider = aws.eu-north-1

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "eu-west-1" {
  count    = contains(var.lambda_regions, "eu-west-1") ? 1 : 0
  provider = aws.eu-west-1

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "eu-west-2" {
  count    = contains(var.lambda_regions, "eu-west-2") ? 1 : 0
  provider = aws.eu-west-2

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "eu-west-3" {
  count    = contains(var.lambda_regions, "eu-west-3") ? 1 : 0
  provider = aws.eu-west-3

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "sa-east-1" {
  count    = contains(var.lambda_regions, "sa-east-1") ? 1 : 0
  provider = aws.sa-east-1

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "us-east-1" {
  count    = contains(var.lambda_regions, "us-east-1") ? 1 : 0
  provider = aws.us-east-1

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "us-east-2" {
  count    = contains(var.lambda_regions, "us-east-2") ? 1 : 0
  provider = aws.us-east-2

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "us-west-1" {
  count    = contains(var.lambda_regions, "us-west-1") ? 1 : 0
  provider = aws.us-west-1

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}

resource "aws_lambda_function" "us-west-2" {
  count    = contains(var.lambda_regions, "us-west-2") ? 1 : 0
  provider = aws.us-west-2

  filename      = "${path.module}/dummy.zip"
  function_name = "${var.name}-${random_id.this.hex}"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}
