# # Test 1: INVALID - Forge + Fusion2 + cli_path
# # This resource should produce an error during plan/apply
# resource "seqera_compute_env" "invalid_forge_fusion2_with_clipath" {
#   workspace_id = var.workspace_id

#   compute_env = {
#     name           = "invalid-forge-fusion2-clipath"
#     description    = "INVALID: Forge + Fusion2 with cli_path set"
#     platform       = "aws-batch"
#     credentials_id = resource.seqera_aws_credential.aws_credential.credentials_id

#     config = {
#       aws_batch = {
#         region          = "us-east-1"
#         work_dir        = var.work_dir
#         fusion2_enabled = true
#         wave_enabled    = true
#         cli_path        = "/usr/local/bin/aws" # This should cause validation error

#         forge = {
#           type                = "EC2"
#           dispose_on_deletion = true
#           min_cpus            = 0
#           max_cpus            = 100
#           instance_types      = ["m5.large"]
#         }
#       }
#     }
#   }
# }

# ## Test 2: INVALID - Fusion2 without Wave
# # This resource should produce an error during plan/apply
# resource "seqera_compute_env" "invalid_fusion2_without_wave" {
#   workspace_id = var.workspace_id

#   compute_env = {
#     name           = "invalid-fusion2-no-wave"
#     description    = "INVALID: Fusion2 enabled without Wave"
#     platform       = "aws-batch"
#     credentials_id = resource.seqera_aws_credential.aws_credential.credentials_id

#     config = {
#       aws_batch = {
#         region          = "us-east-1"
#         work_dir        = var.work_dir
#         fusion2_enabled = true
#         wave_enabled    = false # This should cause validation error

#         forge = {
#           type                = "EC2"
#           dispose_on_deletion = true
#           min_cpus            = 0
#           max_cpus            = 100
#           instance_types      = ["m5.large"]
#         }
#       }
#     }
#   }
# }
