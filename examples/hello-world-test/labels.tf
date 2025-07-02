# resource "seqera_labels" "my_labels" {
#   is_default   = true
#   name         = "terraform-label"
#   resource     = false
#   value        = "label-created-by-terraform"
#   workspace_id = resource.seqera_workspace.my_workspace.id
# }