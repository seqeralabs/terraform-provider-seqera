# Organization and Workspace Outputs
output "org_id" {
  description = "ID of the created organization"
  value       = seqera_orgs.my_org.org_id
}

output "workspace_id" {
  description = "ID of the created workspace"
  value       = seqera_workspace.my_workspace.id
}

output "team_id" {
  description = "ID of the created team"
  value       = seqera_teams.my_teams.team_id
}

# Dataset and Secret Outputs
output "dataset_id" {
  description = "ID of the created dataset"
  value       = seqera_datasets.my_datasets.dataset.id
}

output "pipeline_secret_id" {
  description = "ID of the created pipeline secret"
  value       = seqera_pipeline_secret.my_pipelinesecret.secret_id
}

# AWS Resource Outputs
output "credential_id" {
  description = "ID of the created AWS credential"
  value       = seqera_credential.aws_credential.credentials_id
}

output "compute_env_id" {
  description = "ID of the created compute environment"
  value       = seqera_compute_env.aws_batch_compute_env.compute_env_id
}

output "data_link_id" {
  description = "ID of the created data link"
  value       = seqera_data_link.my_datalink.data_link_id
}

# Pipeline and Workflow Outputs
output "action_id" {
  description = "ID of the created action"
  value       = seqera_action.my_action.action_id
}

output "pipeline_id" {
  description = "ID of the created pipeline"
  value       = seqera_pipeline.hello_world_minimal.pipeline_id
}

output "workflow_id" {
  description = "ID of the launched workflow"
  value       = seqera_workflows.my_workflows.workflow_id
}

output "data_studio_id" {
  description = "ID of the created data studio"
  value       = seqera_studios.my_datastudios.session_id
}

output "label_id" {
  description = "ID of the created label"
  value       = seqera_labels.my_labels.label_id
}