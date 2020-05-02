output "proxy_url" {
  value = {
    http   = "http://${local.proxy_credentials}@${aws_eip.this.public_ip}:${var.proxy_port}"
    socks5 = "socks5://${local.proxy_credentials}@${aws_eip.this.public_ip}:${var.proxy_port}"
  }
}
