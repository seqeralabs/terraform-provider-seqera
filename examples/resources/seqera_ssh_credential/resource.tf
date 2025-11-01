# SSH Credential Resource Examples
#
# IMPORTANT: Credential names must use only letters, numbers, underscores, and hyphens.
# No spaces allowed. Use snake_case (my_credential) or kebab-case (my-credential).

# Variables for sensitive credentials
variable "ssh_private_key" {
  description = "SSH private key (PEM format)"
  type        = string
  sensitive   = true
}

variable "ssh_passphrase" {
  description = "SSH key passphrase (if encrypted)"
  type        = string
  sensitive   = true
  default     = ""
}

# Example: Basic SSH credentials with private key
resource "seqera_ssh_credential" "example" {
  name         = "ssh-main"
  workspace_id = seqera_workspace.main.id

  private_key = var.ssh_private_key
  passphrase  = var.ssh_passphrase  # Optional, only if key is encrypted
}
