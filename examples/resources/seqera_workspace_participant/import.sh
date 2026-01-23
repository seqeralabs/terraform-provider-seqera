#!/bin/bash

#
# Import an existing workspace participant by email.
# Format: org_id/workspace_id/email
#---
terraform import seqera_workspace_participant.user_by_email '12345/67890/user@example.com'

#
# Import an existing workspace participant by member_id.
# Format: org_id/workspace_id/member:member_id
#---
terraform import seqera_workspace_participant.user_by_member_id '12345/67890/member:98765'

#
# Import an existing team participant by team_id.
# Format: org_id/workspace_id/team:team_id
#---
terraform import seqera_workspace_participant.team_access '12345/67890/team:7405043533023'
