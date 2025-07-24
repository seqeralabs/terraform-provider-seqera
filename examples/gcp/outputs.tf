# # Organization and Workspace Outputs
# output "org_id" {
#   description = "ID of the created organization"
#   value       = seqera_orgs.my_org.org_id
# }

# output "workspace_id" {
#   description = "ID of the created workspace"
#   value       = seqera_workspace.my_workspace.id
# }

# # Dataset Outputs
# output "dataset_id" {
#   description = "ID of the created dataset"
#   value       = seqera_datasets.my_datasets.datasets_id
# }

# # GCP Resource Outputs
# output "credential_id" {
#   description = "ID of the created GCP credential"
#   value       = seqera_credential.gcp_credential.credentials_id
# }

# output "compute_env_id" {
#   description = "ID of the created compute environment"
#   value       = seqera_compute_env.gcp_batch_compute_env.compute_env_id
# }

# # Pipeline and Workflow Outputs
# output "pipeline_id" {
#   description = "ID of the created pipeline"
#   value       = seqera_pipeline.hello_world_minimal.pipeline_id
# }

# output "workflow_id" {
#   description = "ID of the launched workflow"
#   value       = seqera_workflows.my_workflows.workflow_id
# }