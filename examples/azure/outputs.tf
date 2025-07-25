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

output "pipeline_secret_id" {
  description = "ID of the created pipeline secret"
  value       = seqera_pipeline_secret.my_pipelinesecret.secret_id
}

output "pipeline_id" {
  description = "ID of the created RNA-seq pipeline"
  value       = seqera_pipeline.rnaseq_pipeline.pipeline_id
}

output "workflow_id" {
  description = "ID of the launched RNA-seq workflow"
  value       = seqera_workflows.rnaseq_workflow.workflow_id
}

output "label_id" {
  description = "ID of the created label"
  value       = seqera_labels.my_labels.label_id
}