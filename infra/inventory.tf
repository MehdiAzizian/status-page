resource "local_file" "inventory" {
  filename = "${path.module}/../ansible/inventory.ini"
  content  = "[web]\n${aws_instance.web.public_ip} ansible_user=ubuntu ansible_ssh_private_key_file=~/.ssh/status-page\n"
}
