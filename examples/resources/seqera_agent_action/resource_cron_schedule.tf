# Runs an agent every weekday at 08:00 Madrid time.
resource "seqera_agent_action" "daily_report" {
  workspace_id = seqera_workspace.main.id
  name         = "daily-run-report"
  source       = "cron"
  agent_id     = seqera_agent.reporter.id

  cron = {
    expression = "0 8 * * 1-5"
    timezone   = "Europe/Madrid"
  }
}
