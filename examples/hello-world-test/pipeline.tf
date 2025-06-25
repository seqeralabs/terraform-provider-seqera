resource "seqera_pipeline" "my_pipeline" {
  compute_env_id = resource.seqera_compute_env.aws_batch_compute_env.compute_env_id
  name                 = "terraform-hello-world"
  config_profiles = [  ]
  config_text        = "...my_config_text..."
  date_created       = "2020-04-27T05:44:01.599Z"
  description        = "...my_description..."
  head_job_cpus      = 4
  head_job_memory_mb = 9

  pipeline             = "https://github.com/nextflow-io/hello"
  pull_latest          = true
  resume               = true
  revision             = "master"
  run_name             = "terraform-hello"
  work_dir     = local.work_dir
  workspace_id = resource.seqera_workspace.my_workspace.id

}

resource "seqera_pipeline" "hello_world_minimal" {
  workspace_id   = resource.seqera_workspace.my_workspace.id
  compute_env_id = resource.seqera_compute_env.aws_batch_compute_env.compute_env_id
  
  name        = "hello-world-minimal"
  description = "Minimal hello world pipeline"
  pipeline    = "https://github.com/nextflow-io/hello"
  revision    = "master"
  work_dir    = local.work_dir
  
  # Basic resource allocation
  head_job_cpus      = 1
  head_job_memory_mb = 2048
}


resource "seqera_pipeline" "test1" {
  workspace_id   = resource.seqera_workspace.my_workspace.id
  compute_env_id = resource.seqera_compute_env.aws_batch_compute_env.compute_env_id
  
  name        = "hello"                                    # 5 chars, alpha only
  description = "Test pipeline 1"
  pipeline    = "https://github.com/nextflow-io/hello"
  revision    = "master"
  work_dir    = local.work_dir
}
