resource "seqera_trace" "my_trace" {
  launch_id    = "...my_launch_id..."
  project_name = "...my_project_name..."
  repository   = "...my_repository..."
  run_name     = "...my_run_name..."
  session_id   = "...my_session_id..."
  workflow_id  = "...my_workflow_id..."
  workspace_id = 4
}