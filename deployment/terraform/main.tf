locals {
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
      --name ${var.name} \
      --env AWS_ACCESS_KEY_ID=${aws_iam_access_key.this.id} \
      --env AWS_SECRET_ACCESS_KEY=${aws_iam_access_key.this.secret} \
      --env LAMBDA_NAME="${var.name}-${random_id.this.hex}" \
      --env LAMBDA_IAM_ROLE_NAME="${aws_iam_role.lambda.name}" \
      --env REGIONS=${join(",", var.lambda_regions)} \
      --env FREQUENCY=${var.lambda_frequency} \
      --env SSH_USER=${var.tunnel_ssh_user} \
      --env SSH_PORT=${var.tunnel_ssh_port} \
      --env MEMORY=${var.lambda_memory} \
      --env LISTENERS=${local.proxy_credentials}@:${var.proxy_port}?dns=${var.proxy_dns} \
      --env DEBUG=${var.app_debug} \
      --env DEBUG_PROXY=${var.proxy_debug} \
      --env BYPASS=${join(",", var.proxy_bypass_domains)} \
      --publish ${var.tunnel_ssh_port}:${var.tunnel_ssh_port} \
      --publish ${var.proxy_port}:${var.proxy_port} \
      ${local.docker_image}:${var.app_version}
  EOF

  stop_awslambdaproxy = "docker rm -f ${var.name}"

  default_subnet = element(tolist(data.aws_subnet_ids.default.ids), 0)
  custom_subnet  = try(aws_subnet.public[0].id, "")
}

resource "random_id" "this" {
  byte_length = 1

  keepers = {
    cidr_block = var.vpc_cidr_block
  }
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
  ami                    = data.aws_ami.this.id
  instance_type          = var.instance_type
  iam_instance_profile   = aws_iam_instance_profile.this.name
  vpc_security_group_ids = [for i in aws_security_group.this : i.id]
  subnet_id              = var.create_vpc ? local.custom_subnet : local.default_subnet
  key_name               = aws_key_pair.ec2.key_name

  provisioner "remote-exec" {
    inline = local.install_docker

    connection {
      host        = aws_instance.this.public_ip
      user        = "ec2-user"
      private_key = tls_private_key.ec2.private_key_pem
    }
  }

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_eip" "this" {
  vpc = true

  tags = {
    Name      = var.name
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
    Name      = var.name
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
  for_each = toset(var.lambda_regions)

  name   = "${var.name}-${each.value}-${random_id.this.hex}"
  vpc_id = var.create_vpc ? try(aws_vpc.this[0].id, "") : data.aws_vpc.default.id

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
    cidr_blocks = data.aws_ip_ranges.lambda[each.value].cidr_blocks
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
    Name      = var.name
    Workspace = terraform.workspace
  }
}

resource "aws_iam_instance_profile" "this" {
  name = "${var.name}-instance-profile-${random_id.this.hex}"
  role = aws_iam_role.profile.name
}

resource "aws_iam_role" "profile" {
  name               = "${var.name}-profile-${random_id.this.hex}"
  assume_role_policy = data.aws_iam_policy_document.profile_sts.json

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }
}

resource "aws_iam_role_policy" "profile" {
  policy = data.aws_iam_policy_document.user.json
  role   = aws_iam_role.profile.id
}

resource "aws_iam_role" "lambda" {
  name               = "${var.name}-lambda-${random_id.this.hex}"
  assume_role_policy = data.aws_iam_policy_document.lambda_sts.json

  tags = {
    Name      = var.name
    Workspace = terraform.workspace
  }
}

resource "aws_iam_role_policy" "lambda" {
  policy = data.aws_iam_policy_document.lambda.json
  role   = aws_iam_role.lambda.id
}

resource "aws_iam_user" "this" {
  name = "${var.name}-user-${random_id.this.hex}"

  tags = {
    Name      = var.name
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
