resource "seqera_pipeline" "my_pipeline" {
  description = "...my_description..."
  icon        = "...my_icon..."
  label_ids = [
    7
  ]
  launch = {
    commit_id      = "...my_commit_id..."
    compute_env_id = "4g09tT4pW4JFUvXTHdB6zP"
    config_profiles = [
      "docker",
      "aws",
    ]
    config_text        = "process {\n  executor = 'awsbatch'\n  queue = 'my-queue'\n}\n"
    entry_name         = "main.nf"
    head_job_cpus      = 2
    head_job_memory_mb = 4096
    label_ids = [
      1001,
      1002,
      1003,
    ]
    main_script     = "main.nf"
    params_text     = "{\n  \"input\": \"s3://my-bucket/input.csv\",\n  \"output_dir\": \"s3://my-bucket/results\"\n}\n"
    pipeline        = "https://github.com/nextflow-io/hello"
    post_run_script = "#!/bin/bash\necho \"Workflow completed\"\naws s3 sync ./results s3://my-bucket/results\n"
    pre_run_script  = "#!/bin/bash\necho \"Starting workflow execution\"\naws s3 sync s3://my-bucket/data ./data\n"
    pull_latest     = true
    resume          = true
    revision        = "main"
    run_name        = "nextflow-hello"
    schema_name     = "nextflow_schema.json"
    stub_run        = false
    tower_config    = "...my_tower_config..."
    user_secrets = [
      "MY_API_KEY",
      "DATABASE_PASSWORD",
    ]
    work_dir = "s3://my-bucket/work"
    workspace_secrets = [
      "WORKSPACE_TOKEN",
      "SHARED_CREDENTIALS",
    ]
  }
  name         = "rna-seq-analysis"
  workspace_id = 3
}