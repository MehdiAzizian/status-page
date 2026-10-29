# 1. Role: identity for the server. Who may use it = EC2.
resource "aws_iam_role" "web" {
  name = "status-page-web"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

# 2. Permission: this role may talk to SSM.
resource "aws_iam_role_policy_attachment" "ssm" {
  role       = aws_iam_role.web.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

# 3. Connector: attaches the role to a server.
resource "aws_iam_instance_profile" "web" {
  name = "status-page-web"
  role = aws_iam_role.web.name
}
