package nextflow

import (
	"strings"
)

// buildExecutorBlock builds the executor configuration block
func buildExecutorBlock(executor ExecutorModel) string {
	if isExecutorEmpty(executor) {
		return ""
	}

	var b strings.Builder
	b.WriteString("executor {\n")

	if !executor.QueueSize.IsNull() && !executor.QueueSize.IsUnknown() {
		b.WriteString(writeNumberField("queue_size", executor.QueueSize.ValueInt64()))
	}
	if !executor.PollInterval.IsNull() && !executor.PollInterval.IsUnknown() {
		b.WriteString(writeStringField("poll_interval", executor.PollInterval.ValueString()))
	}
	if !executor.QueueStatInterval.IsNull() && !executor.QueueStatInterval.IsUnknown() {
		b.WriteString(writeStringField("queue_stat_interval", executor.QueueStatInterval.ValueString()))
	}
	if !executor.SubmitRateLimit.IsNull() && !executor.SubmitRateLimit.IsUnknown() {
		b.WriteString(writeStringField("submit_rate_limit", executor.SubmitRateLimit.ValueString()))
	}
	if !executor.ExitReadTimeout.IsNull() && !executor.ExitReadTimeout.IsUnknown() {
		b.WriteString(writeStringField("exit_read_timeout", executor.ExitReadTimeout.ValueString()))
	}

	b.WriteString("}\n\n")
	return b.String()
}

// isExecutorEmpty checks if the executor block has any non-null values
func isExecutorEmpty(executor ExecutorModel) bool {
	return (executor.QueueSize.IsNull() || executor.QueueSize.IsUnknown()) &&
		(executor.PollInterval.IsNull() || executor.PollInterval.IsUnknown()) &&
		(executor.QueueStatInterval.IsNull() || executor.QueueStatInterval.IsUnknown()) &&
		(executor.SubmitRateLimit.IsNull() || executor.SubmitRateLimit.IsUnknown()) &&
		(executor.ExitReadTimeout.IsNull() || executor.ExitReadTimeout.IsUnknown())
}
