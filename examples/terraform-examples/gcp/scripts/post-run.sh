#!/bin/bash

# Post-run script for Seqera Platform workflows
# This script runs after workflow execution completes

set -euo pipefail

echo "Workflow execution completed!"
echo "Post-run script executed at: $(date)"

# Cleanup operations
echo "Performing cleanup operations..."

# Add any custom cleanup commands here
# Examples:
# - Archive or move output files
# - Clean up temporary directories
# - Send notifications
# - Update databases or logs
# - Generate reports

echo "Post-run cleanup completed successfully"