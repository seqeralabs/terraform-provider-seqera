resource "seqera_workspace" "my_workspace" {
  description = "Workspace for bioinformatics research projects"
  full_name   = "My Research Workspace"
  id          = 1
  name        = "my-research-workspace"
  org_id      = 7
  visibility  = "SHARED"
}