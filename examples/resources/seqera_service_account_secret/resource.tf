resource "seqera_service_account" "jobs" {
  org_id      = var.org_id
  name        = "scheduled-jobs"
  description = "Identity for scheduled jobs"
}

# Rotate the secret every 300 days, well before its 365-day expiry.
resource "time_rotating" "jobs_secret" {
  rotation_days = 300
}

resource "seqera_service_account_secret" "jobs" {
  org_id             = var.org_id
  service_account_id = seqera_service_account.jobs.id
  expiry_days        = 365

  lifecycle {
    # Issue the new secret before revoking the old one, so there is no gap.
    create_before_destroy = true
    replace_triggered_by  = [time_rotating.jobs_secret]
  }
}

# Hand the credentials to the jobs through a secrets manager.
resource "aws_secretsmanager_secret_version" "jobs" {
  secret_id = aws_secretsmanager_secret.jobs.id
  secret_string = jsonencode({
    client_id     = seqera_service_account_secret.jobs.client_id
    client_secret = seqera_service_account_secret.jobs.secret
  })
}
