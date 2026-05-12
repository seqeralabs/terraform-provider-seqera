# LARGE Seqera Managed Compute environment with a custom work directory suffix.
# 28-day automatic intermediary cleanup is on by default.
resource "seqera_managed_compute_ce" "large" {
  name                  = "seqera-cloud-large"
  workspace_id          = data.seqera_workspace.main.id
  region                = "eu-west-1"
  instance_size         = "LARGE"
  work_dir              = "rnaseq/work"
  data_retention_policy = true
}
