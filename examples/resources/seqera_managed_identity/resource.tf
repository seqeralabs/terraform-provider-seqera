resource "seqera_managed_identity" "my_managedidentity" {
  checked             = false
  host_name           = "slurm.example.com"
  managed_identity_id = 1
  name                = "my-slurm-cluster"
  org_id              = 8
  platform            = "altair-platform"
  port                = 22
}