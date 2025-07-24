output "org_id" {
  description = "ID of the created organization"
  value       = seqera_orgs.my_org.org_id
}

output "workspace_id" {
  description = "ID of the created workspace"
  value       = seqera_workspace.my_workspace.id
}

output "dataset_id" {
  description = "ID of the created dataset"
  value       = seqera_datasets.my_datasets.dataset.id
}

output "credential_id" {
  description = "ID of the created Azure credential"
  value       = seqera_credential.azure_credential.credentials_id
}

output "compute_env_id" {
  description = "ID of the created compute environment"
  value       = seqera_compute_env.azure_batch_compute_env.compute_env_id
}

output "pipeline_id" {
  description = "ID of the created pipeline"
  value       = seqera_pipeline.hello_world_minimal.pipeline_id
}

output "workflow_id" {
  description = "ID of the launched workflow"
  value       = seqera_workflows.my_workflows.workflow_id
}