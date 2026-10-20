# Use one of the two configurations below per workspace: exactly one resource
# may manage a given workspace's lineage settings, so the AUTO and MANUAL
# examples use different workspaces.
# var.* and seqera_aws_credential.main are placeholders for your own values.
#
# AUTO provisioning (default): Platform creates the lineage bucket, the
# notification topic and the subscription. Do not set store_name or topic_arn.
#
# credentials_id and region are locked once AUTO provisioning has completed; changing
# them afterwards is rejected by the Platform.
#
# Destroying this resource disables lineage for the workspace. The AWS
# infrastructure that was created is left in place.
resource "seqera_workspace_lineage_settings" "auto" {
  org_id            = var.auto_org_id
  workspace_id      = var.auto_workspace_id
  provisioning_mode = "AUTO"

  credentials_id  = seqera_aws_credential.main.id
  region          = "eu-west-1"
  lineage_default = true
}

# MANUAL provisioning: you create the bucket and the SNS topic, and subscribe
# the Platform webhook to the topic. store_name and topic_arn are required.
resource "aws_sns_topic" "lineage" {
  name = "seqera-lineage-events"
}

resource "seqera_workspace_lineage_settings" "manual" {
  org_id            = var.manual_org_id
  workspace_id      = var.manual_workspace_id
  provisioning_mode = "MANUAL"

  credentials_id = seqera_aws_credential.main.id
  region         = "eu-west-1"
  store_name     = "my-existing-lineage-bucket"
  topic_arn      = aws_sns_topic.lineage.arn
}

# The subscription is confirmed by the Platform. event_delivery stays
# PENDING_CONFIRMATION until the confirmation completes.
resource "aws_sns_topic_subscription" "lineage" {
  topic_arn = aws_sns_topic.lineage.arn
  protocol  = "https"
  endpoint  = seqera_workspace_lineage_settings.manual.webhook_url
}
