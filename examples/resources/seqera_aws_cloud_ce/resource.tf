# Look up the target organization and workspace by name.
data "seqera_organization" "main" {
  name = "my-organization"
}

data "seqera_workspace" "main" {
  org_id = data.seqera_organization.main.org_id
  name   = "my-workspace"
}

# Minimal AWS Cloud compute environment (Classic mode).
# Seqera picks the worker fleet automatically.
#
# If you set `allow_buckets` explicitly, include the `work_dir` URI as the
# trailing entry — Seqera Forge implicitly appends it at CE-create time, and
# omitting it produces a forced-replacement diff on subsequent plans.
resource "seqera_aws_cloud_ce" "classic" {
  name           = "aws-cloud-classic"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region   = "us-west-1"
    work_dir = "s3://my-bucket/work"
  }
}

# AWS Cloud compute environment with explicit networking and an encrypted
# boot volume.
resource "seqera_aws_cloud_ce" "networked" {
  name           = "aws-cloud-networked"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region   = "us-west-1"
    work_dir = "s3://my-bucket/work"

    # Networking. Use subnet_ids (a list) rather than the deprecated subnet_id.
    vpc_id     = "vpc-12345678"
    subnet_ids = ["subnet-12345678", "subnet-87654321"]

    # Encrypt the boot EBS volume. ebs_kms_key_id requires ebs_encrypted = true;
    # omit it to use the account/region default EBS encryption key.
    ebs_encrypted  = true
    ebs_kms_key_id = "arn:aws:kms:us-west-1:123456789012:key/12345678-90ab-cdef-1234-567890abcdef"
  }
}
