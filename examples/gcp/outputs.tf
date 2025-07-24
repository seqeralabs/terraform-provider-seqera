# Resource Outputs
output "credential_id" {
  description = "ID of the created GCP credential"
  value       = seqera_credential.gcp_credential.credentials_id
}

output "compute_env_id" {
  description = "ID of the created compute environment"
  value       = seqera_compute_env.gcp_batch_compute_env.compute_env_id
}

output "pipeline_id" {
  description = "ID of the created pipeline"
  value       = seqera_pipeline.hello_world_minimal.pipeline_id
}

output "workflow_id" {
  description = "ID of the launched workflow"
  value       = seqera_workflows.my_workflows.workflow_id
}