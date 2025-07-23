resource "seqera_workflows" "my_workflows" {
  compute_env_id = "ce_67890fghij"
  config_profiles = [
    "..."
  ]
  config_text        = "process {\n  executor = 'awsbatch'\n  queue = 'my-queue'\n}\n"
  date_created       = "2024-07-23T10:30:00Z"
  entry_name         = "MAIN"
  force              = false
  head_job_cpus      = 2
  head_job_memory_mb = 4096
  label_ids = [
    6
  ]
  launch_container     = "quay.io/seqeralabs/nf-launcher:latest"
  main_script          = "main.nf"
  optimization_id      = "opt_98765zyxwv"
  optimization_targets = "cost,time"
  params_text          = "{\n  \"input\": \"s3://my-bucket/input.csv\",\n  \"output_dir\": \"s3://my-bucket/results\",\n  \"max_cpus\": 16\n}\n"
  pipeline             = "https://github.com/nextflow-io/hello"
  post_run_script      = "#!/bin/bash\necho \"Workflow completed\"\naws s3 sync ./results s3://my-bucket/results\n"
  pre_run_script       = "#!/bin/bash\necho \"Starting workflow execution\"\naws s3 sync s3://my-bucket/data ./data\n"
  pull_latest          = false
  resume               = true
  revision             = "main"
  run_name             = "my-workflow-run-2024"
  schema_name          = "nextflow_schema.json"
  session_id           = "...my_session_id..."
  source_workspace_id  = 2
  stub_run             = false
  tower_config         = "tower {\n  accessToken = '$TOWER_ACCESS_TOKEN'\n  workspaceId = 'my-workspace'\n}\n"
  user_secrets = [
    "..."
  ]
  work_dir     = "s3://my-bucket/work"
  workspace_id = 10
  workspace_secrets = [
    "..."
  ]
}
