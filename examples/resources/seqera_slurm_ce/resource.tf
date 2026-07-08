# Slurm HPC compute environment.
#
# Seqera connects to the cluster's login/head node over SSH using the
# referenced seqera_ssh_credential, then submits the Nextflow head job to
# Slurm. work_dir must be a path on a filesystem shared across the cluster
# nodes; it is required and force-new — changing it replaces the CE.
#
# Config fields are hoisted to the resource root (no nested `config` block).
resource "seqera_slurm_ce" "hpc" {
  name           = "slurm-hpc"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_ssh_credential.hpc.credentials_id

  work_dir   = "/scratch/users/me/seqera/work"
  launch_dir = "/scratch/users/me/seqera/launch"
  user_name  = "me"
  host_name  = "login.hpc.example.org"

  # Options passed to the Slurm head job submission (sbatch).
  head_job_options = "-t 72:00:00 --cpus-per-task=2 --mem-per-cpu=8G"

  pre_run_script = trimspace(<<-EOT
    module load nextflow
    source activate /home/me/.conda/envs/nextflow/
  EOT
  )
}
