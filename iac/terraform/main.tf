locals {
  app_name = "awslambdaproxy"

  docker_image = "vdan/awslambdaproxy"

  proxy_credentials = var.proxy_credentials == null ? "${random_string.this.result}:${random_password.this.result}" : var.proxy_credentials

  install_docker = [
    "sudo yum update -y",
    "sudo amazon-linux-extras install docker -y",
    "sudo service docker start",
    "sudo usermod -a -G docker ec2-user",
  ]

  start_awslambdaproxy = <<-EOF
    docker run --detach --restart=always \
      --name ${local.app_name} \
      --env AWS_ACCESS_KEY_ID=${aws_iam_access_key.this.id} \
      --env AWS_SECRET_ACCESS_KEY=${aws_iam_access_key.this.secret} \
      --env REGIONS=${join(",", var.lambda_regions)} \
      --env FREQUENCY=${var.lambda_frequency} \
      --env SSH_USER=${var.tunnel_ssh_user} \
      --env SSH_PORT=${var.tunnel_ssh_port} \
      --env MEMORY=${var.lambda_memory} \
      --env LISTENERS=${local.proxy_credentials}@:${var.proxy_port} \
      --env DEBUG_PROXY=${var.proxy_debug} \
      --publish ${var.tunnel_ssh_port}:${var.tunnel_ssh_port} \
      --publish ${var.proxy_port}:${var.proxy_port} \
      ${local.docker_image}:${var.app_version}
  EOF

  stop_awslambdaproxy = "docker rm -f ${local.app_name}"
}

resource "random_id" "this" {
  byte_length = 2
}

resource "random_string" "this" {
  length  = 8
  special = false
}

resource "random_password" "this" {
  length  = 16
  special = false
}

resource "aws_instance" "this" {
  ami                  = data.aws_ami.this.id
  instance_type        = var.instance_type
  availability_zone    = data.aws_availability_zones.this.names[0]
  iam_instance_profile = aws_iam_instance_profile.this.name
  security_groups      = [data.aws_security_group.default.name, aws_security_group.this.name]
  key_name             = aws_key_pair.ec2.key_name

  provisioner "remote-exec" {
    inline = local.install_docker

    connection {
      host        = aws_instance.this.public_ip
      user        = "ec2-user"
      private_key = tls_private_key.ec2.private_key_pem
    }
  }

  tags = {
    Name      = "${var.name}-${random_id.this.hex}"
    Workspace = terraform.workspace
  }

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [aws_lambda_function.this]
}

resource "aws_eip" "this" {
  vpc = true

  tags = {
    Name      = "${var.name}-${random_id.this.hex}"
    Workspace = terraform.workspace
  }
}

resource "aws_eip_association" "this" {
  allocation_id = aws_eip.this.id
  instance_id   = aws_instance.this.id
}

resource "null_resource" "start_awslambdaproxy" {
  provisioner "remote-exec" {
    inline = [local.start_awslambdaproxy]

    connection {
      host        = aws_eip_association.this.public_ip
      user        = "ec2-user"
      private_key = tls_private_key.ec2.private_key_pem
    }
  }
}

resource "null_resource" "restart_awslambdaproxy" {
  triggers = {
    start_awslambdaproxy = local.start_awslambdaproxy
  }

  provisioner "remote-exec" {
    inline = [local.stop_awslambdaproxy, local.start_awslambdaproxy]

    connection {
      host        = aws_eip_association.this.public_ip
      user        = "ec2-user"
      private_key = tls_private_key.ec2.private_key_pem
    }
  }

  depends_on = [null_resource.start_awslambdaproxy]
}

resource "tls_private_key" "ec2" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "ec2" {
  key_name   = "${var.name}-${random_id.this.hex}"
  public_key = tls_private_key.ec2.public_key_openssh
}

resource "aws_secretsmanager_secret" "this" {
  name = "${var.name}-${random_id.this.hex}"

  tags = {
    Name      = "${var.name}-${random_id.this.hex}"
    Workspace = terraform.workspace
  }
}

resource "aws_secretsmanager_secret_version" "this" {
  secret_id = aws_secretsmanager_secret.this.id
  secret_string = jsonencode({
    public_key  = tls_private_key.ec2.public_key_openssh
    private_key = tls_private_key.ec2.private_key_pem
  })
}

resource "aws_security_group" "this" {
  name = "${var.name}-${random_id.this.hex}"

  dynamic "ingress" {
    for_each = var.ssh_cidr_blocks
    content {
      from_port   = 22
      protocol    = "tcp"
      to_port     = 22
      cidr_blocks = [ingress.value]
    }
  }

  ingress {
    from_port   = 22
    protocol    = "tcp"
    to_port     = 22
    cidr_blocks = ["${chomp(data.http.current_ip.body)}/32"]
  }

  ingress {
    from_port   = var.tunnel_ssh_port
    protocol    = "tcp"
    to_port     = var.tunnel_ssh_port
    cidr_blocks = data.aws_ip_ranges.lambda.cidr_blocks
  }

  ingress {
    from_port   = var.proxy_port
    protocol    = "tcp"
    to_port     = var.proxy_port
    cidr_blocks = var.proxy_cidr_blocks
  }

  egress {
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name      = "${var.name}-${random_id.this.hex}"
    Workspace = terraform.workspace
  }
}

resource "aws_iam_instance_profile" "this" {
  name = "${var.name}-instance-profile-${random_id.this.hex}"
  role = aws_iam_role.profile.name
}

resource "aws_iam_role" "profile" {
  name               = "${var.name}-role-${random_id.this.hex}"
  assume_role_policy = data.aws_iam_policy_document.profile_sts.json

  tags = {
    Name      = "${var.name}-${random_id.this.hex}"
    Workspace = terraform.workspace
  }
}

resource "aws_iam_role_policy" "profile" {
  policy = data.aws_iam_policy_document.user.json
  role   = aws_iam_role.profile.id
}

resource "aws_iam_role" "lambda" {
  name               = "awslambdaproxy-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_sts.json

  tags = {
    Name      = "${var.name}-${random_id.this.hex}"
    Workspace = terraform.workspace
  }
}

resource "aws_iam_role_policy_attachment" "test-attach" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

resource "aws_iam_role_policy" "lambda" {
  policy = data.aws_iam_policy_document.lambda.json
  role   = aws_iam_role.lambda.id
}

resource "aws_iam_user" "this" {
  name = "${var.name}-user-${random_id.this.hex}"

  tags = {
    Name      = "${var.name}-${random_id.this.hex}"
    Workspace = terraform.workspace
  }
}

resource "aws_iam_user_policy" "this" {
  name   = "${var.name}-user-${random_id.this.hex}"
  user   = aws_iam_user.this.name
  policy = data.aws_iam_policy_document.user.json
}

resource "aws_iam_access_key" "this" {
  user = aws_iam_user.this.name
}

resource "aws_lambda_function" "this" {
  for_each = toset(var.lambda_regions)

  filename      = "${path.module}/function.zip"
  function_name = "awslambdaproxy"
  handler       = "main"
  role          = aws_iam_role.lambda.arn
  runtime       = "go1.x"

  tags = {
    Name      = "${var.name}-${random_id.this.hex}"
    Workspace = terraform.workspace
  }

  lifecycle {
    ignore_changes = [timeout]
  }
}
