resource "seqera_workspace" "my_workspace" {
  description = "Workspace for genomics research projects and computational biology workflows"
  full_name   = "Genomics Research Workspace"
  name        = "genomics-research"
  org_id      = 7
  visibility  = "SHARED"
}