# Runs an agent when the marker file appears in the watched data link.
resource "seqera_agent_action" "on_upload" {
  workspace_id = seqera_workspace.main.id
  name         = "check-new-uploads"
  source       = "bucket"
  agent_id     = seqera_agent.qc.id

  bucket = {
    data_link_id = seqera_data_link.inputs.data_link_id
    marker_file  = "ready.txt"
  }
}
