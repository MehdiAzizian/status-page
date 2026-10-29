# GitHub's identity: an IAM user (OIDC was blocked)
resource "aws_iam_user" "github" {
  name = "status-page-github"
}

# Permission: only send commands to MY server
resource "aws_iam_user_policy" "github_ssm" {
  user = aws_iam_user.github.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "ssm:SendCommand"
        Resource = [
          aws_instance.web.arn,
          "arn:aws:ssm:eu-north-1::document/AWS-RunShellScript"
        ]
      },
      {
        Effect   = "Allow"
        Action   = "ssm:GetCommandInvocation"
        Resource = "*"
      }
    ]
  })
}
