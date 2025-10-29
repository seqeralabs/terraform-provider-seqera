# Seqera SSH Credentials Examples
#
# SSH credentials store SSH private keys for secure access to remote compute
# environments and resources within the Seqera Platform workflows.
#
# SECURITY BEST PRACTICES:
# - Never hardcode SSH private keys in Terraform files
# - Use Terraform variables marked as sensitive
# - Store actual keys in secure secret management systems
# - Use strong passphrases for SSH keys
# - Regularly rotate SSH keys
# - Use separate keys for different environments
# - Restrict key permissions (chmod 600)
# - Use ed25519 keys when possible (more secure than RSA)

# Variable declarations for sensitive SSH credentials
variable "ssh_private_key" {
  description = "SSH private key content"
  type        = string
  sensitive   = true
}

variable "ssh_passphrase" {
  description = "SSH private key passphrase"
  type        = string
  sensitive   = true
  default     = "" # Empty if no passphrase
}

# =============================================================================
# Example 1: Basic SSH Credentials Without Passphrase
# =============================================================================
# Basic configuration with SSH private key (no passphrase protection)

resource "seqera_ssh_credential" "basic" {
  name         = "ssh-main"
  workspace_id = seqera_workspace.main.id

  private_key = var.ssh_private_key
}

# =============================================================================
# Example 2: SSH Credentials With Passphrase (Recommended)
# =============================================================================
# SSH key protected with a passphrase for enhanced security

resource "seqera_ssh_credential" "with_passphrase" {
  name         = "ssh-secure"
  workspace_id = seqera_workspace.main.id

  private_key = var.ssh_private_key
  passphrase  = var.ssh_passphrase
}

# =============================================================================
# Example 3: Reading SSH Key from File
# =============================================================================
# Load SSH private key from a local file

resource "seqera_ssh_credential" "from_file" {
  name         = "ssh-from-file"
  workspace_id = seqera_workspace.main.id

  private_key = file("~/.ssh/id_ed25519")
  passphrase  = var.ssh_passphrase
}

# =============================================================================
# Example 4: Multiple SSH Keys for Different Environments
# =============================================================================

locals {
  ssh_keys = {
    "dev" = {
      key        = var.ssh_dev_key
      passphrase = var.ssh_dev_passphrase
    }
    "staging" = {
      key        = var.ssh_staging_key
      passphrase = var.ssh_staging_passphrase
    }
    "prod" = {
      key        = var.ssh_prod_key
      passphrase = var.ssh_prod_passphrase
    }
  }
}

resource "seqera_ssh_credential" "by_environment" {
  for_each = local.ssh_keys

  name         = "ssh-${each.key}"
  workspace_id = seqera_workspace.main.id
  private_key  = each.value.key
  passphrase   = each.value.passphrase
}

# =============================================================================
# Example 5: SSH Credentials for Different Compute Environments
# =============================================================================

resource "seqera_ssh_credential" "hpc_cluster" {
  name         = "ssh-hpc-cluster"
  workspace_id = seqera_workspace.main.id
  private_key  = var.hpc_ssh_key
  passphrase   = var.hpc_ssh_passphrase
}

resource "seqera_ssh_credential" "cloud_vms" {
  name         = "ssh-cloud-vms"
  workspace_id = seqera_workspace.main.id
  private_key  = var.cloud_ssh_key
}

# =============================================================================
# Example 6: Generating SSH Keys
# =============================================================================
# To generate a new SSH key pair:
#
# Option 1: Ed25519 Key (Recommended - More Secure):
# ssh-keygen -t ed25519 -C "seqera-platform" -f ~/.ssh/seqera_ed25519
#
# Option 2: RSA Key (Traditional):
# ssh-keygen -t rsa -b 4096 -C "seqera-platform" -f ~/.ssh/seqera_rsa
#
# With passphrase (recommended):
# ssh-keygen -t ed25519 -C "seqera-platform" -f ~/.ssh/seqera_ed25519 -N "your-passphrase"
#
# Without passphrase (less secure):
# ssh-keygen -t ed25519 -C "seqera-platform" -f ~/.ssh/seqera_ed25519 -N ""
#
# Key files created:
# - Private key: ~/.ssh/seqera_ed25519 (NEVER share this)
# - Public key: ~/.ssh/seqera_ed25519.pub (share with systems you want to access)

# =============================================================================
# Example 7: Using SSH Keys with Compute Environments
# =============================================================================

resource "seqera_ssh_credential" "compute_env_key" {
  name         = "ssh-compute"
  workspace_id = seqera_workspace.main.id
  private_key  = var.ssh_private_key
  passphrase   = var.ssh_passphrase
}

# Example: SSH credentials can be used with compute environments
# that require SSH access (e.g., HPC clusters, on-premises systems)

# =============================================================================
# Example 8: AWS Secrets Manager Integration
# =============================================================================
# Store SSH keys securely in AWS Secrets Manager

data "aws_secretsmanager_secret_version" "ssh_key" {
  secret_id = "seqera/ssh/private-key"
}

locals {
  ssh_secret = jsondecode(data.aws_secretsmanager_secret_version.ssh_key.secret_string)
}

resource "seqera_ssh_credential" "from_secrets_manager" {
  name         = "ssh-from-asm"
  workspace_id = seqera_workspace.main.id
  private_key  = local.ssh_secret.private_key
  passphrase   = local.ssh_secret.passphrase
}

# To store in AWS Secrets Manager:
# aws secretsmanager create-secret \
#   --name seqera/ssh/private-key \
#   --secret-string '{"private_key":"-----BEGIN OPENSSH PRIVATE KEY-----\n...","passphrase":"your-passphrase"}'

# =============================================================================
# Example 9: HashiCorp Vault Integration
# =============================================================================
# Retrieve SSH keys from HashiCorp Vault

data "vault_generic_secret" "ssh_key" {
  path = "secret/seqera/ssh-key"
}

resource "seqera_ssh_credential" "from_vault" {
  name         = "ssh-from-vault"
  workspace_id = seqera_workspace.main.id
  private_key  = data.vault_generic_secret.ssh_key.data["private_key"]
  passphrase   = data.vault_generic_secret.ssh_key.data["passphrase"]
}

# =============================================================================
# Example 10: Key Rotation Strategy
# =============================================================================
# Implement SSH key rotation

resource "seqera_ssh_credential" "rotated" {
  name         = "ssh-rotated"
  workspace_id = seqera_workspace.main.id
  private_key  = var.ssh_private_key
  passphrase   = var.ssh_passphrase

  # Key rotation steps:
  # 1. Generate new SSH key pair
  # 2. Deploy new public key to target systems
  # 3. Update Terraform variable with new private key
  # 4. Run terraform apply to update credentials
  # 5. Verify connectivity with new key
  # 6. Remove old public key from target systems
}

# =============================================================================
# Example 11: SSH Key Format Examples
# =============================================================================
# SSH private keys can be in different formats:

# Ed25519 format (recommended):
# -----BEGIN OPENSSH PRIVATE KEY-----
# b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
# ...
# -----END OPENSSH PRIVATE KEY-----

# RSA format (traditional):
# -----BEGIN RSA PRIVATE KEY-----
# MIIEpAIBAAKCAQEA...
# ...
# -----END RSA PRIVATE KEY-----

# New OpenSSH format (most common):
# -----BEGIN OPENSSH PRIVATE KEY-----
# b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABlwAAAAdzc2gtcn
# ...
# -----END OPENSSH PRIVATE KEY-----

# =============================================================================
# Example 12: Converting Key Formats
# =============================================================================
# To convert between key formats:
#
# Convert to OpenSSH format:
# ssh-keygen -p -m RFC4716 -f ~/.ssh/id_rsa
#
# Convert to PEM format:
# ssh-keygen -p -m PEM -f ~/.ssh/id_rsa
#
# Convert public key to different format:
# ssh-keygen -e -f ~/.ssh/id_rsa.pub

# =============================================================================
# Example 13: Deploying Public Keys to Target Systems
# =============================================================================
# After creating SSH credentials in Seqera, deploy the public key to target systems:
#
# Manual deployment:
# ssh-copy-id -i ~/.ssh/seqera_ed25519.pub user@remote-host
#
# Or manually add to authorized_keys:
# cat ~/.ssh/seqera_ed25519.pub | ssh user@remote-host "cat >> ~/.ssh/authorized_keys"
#
# Set correct permissions:
# ssh user@remote-host "chmod 700 ~/.ssh && chmod 600 ~/.ssh/authorized_keys"

# =============================================================================
# SECURITY RECOMMENDATIONS
# =============================================================================
#
# 1. Always use passphrases for SSH keys
# 2. Use Ed25519 keys (more secure and faster than RSA)
# 3. Use minimum RSA key size of 4096 bits if Ed25519 is not available
# 4. Store private keys in secure secret management systems
# 5. Never commit private keys to version control
# 6. Set correct file permissions (600 for private keys)
# 7. Use separate keys for different environments and purposes
# 8. Implement key rotation policies (e.g., every 90 days)
# 9. Remove old public keys from authorized_keys after rotation
# 10. Monitor SSH key usage through system logs
# 11. Disable root login and password authentication on target systems
# 12. Use SSH certificates for large-scale deployments
# 13. Implement audit logging for SSH key usage
# 14. Use hardware security modules (HSM) for critical keys
# 15. Regularly audit and revoke unused keys

# =============================================================================
# Testing SSH Connectivity
# =============================================================================
# After deploying credentials and public keys, test connectivity:
#
# Test basic connectivity:
# ssh -i ~/.ssh/seqera_ed25519 user@remote-host
#
# Test with verbose output (for troubleshooting):
# ssh -vvv -i ~/.ssh/seqera_ed25519 user@remote-host
#
# Test key authentication specifically:
# ssh -i ~/.ssh/seqera_ed25519 -o PreferredAuthentications=publickey user@remote-host
#
# Check authorized_keys on remote:
# ssh user@remote-host "cat ~/.ssh/authorized_keys"
