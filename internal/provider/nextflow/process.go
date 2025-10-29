package nextflow

import (
	"fmt"
	"strings"
)

// buildProcessBlock builds the process configuration block
func buildProcessBlock(process ProcessModel) string {
	if isProcessEmpty(process) {
		return ""
	}

	var b strings.Builder
	b.WriteString("process {\n")

	// Basic fields
	if !process.Executor.IsNull() && !process.Executor.IsUnknown() {
		b.WriteString(writeStringField("executor", process.Executor.ValueString()))
	}
	if !process.Queue.IsNull() && !process.Queue.IsUnknown() {
		b.WriteString(writeStringField("queue", process.Queue.ValueString()))
	}
	if !process.Cpus.IsNull() && !process.Cpus.IsUnknown() {
		b.WriteString(writeNumberField("cpus", process.Cpus.ValueInt64()))
	}
	if !process.Memory.IsNull() && !process.Memory.IsUnknown() {
		b.WriteString(writeStringField("memory", process.Memory.ValueString()))
	}
	if !process.Time.IsNull() && !process.Time.IsUnknown() {
		b.WriteString(writeStringField("time", process.Time.ValueString()))
	}
	if !process.Disk.IsNull() && !process.Disk.IsUnknown() {
		b.WriteString(writeStringField("disk", process.Disk.ValueString()))
	}
	if !process.ErrorStrategy.IsNull() && !process.ErrorStrategy.IsUnknown() {
		b.WriteString(writeStringField("error_strategy", process.ErrorStrategy.ValueString()))
	}
	if !process.MaxRetries.IsNull() && !process.MaxRetries.IsUnknown() {
		b.WriteString(writeNumberField("max_retries", process.MaxRetries.ValueInt64()))
	}
	if !process.MaxErrors.IsNull() && !process.MaxErrors.IsUnknown() {
		b.WriteString(writeNumberField("max_errors", process.MaxErrors.ValueInt64()))
	}
	if !process.ClusterOptions.IsNull() && !process.ClusterOptions.IsUnknown() {
		b.WriteString(writeStringField("cluster_options", process.ClusterOptions.ValueString()))
	}
	if !process.Container.IsNull() && !process.Container.IsUnknown() {
		b.WriteString(writeStringField("container", process.Container.ValueString()))
	}
	if !process.ContainerOptions.IsNull() && !process.ContainerOptions.IsUnknown() {
		b.WriteString(writeStringField("container_options", process.ContainerOptions.ValueString()))
	}

	// withLabel blocks
	for _, label := range process.WithLabel {
		if !label.Name.IsNull() && !label.Name.IsUnknown() {
			b.WriteString("\n")
			b.WriteString(buildWithLabel(label))
		}
	}

	// withName blocks
	for _, name := range process.WithName {
		if !name.Pattern.IsNull() && !name.Pattern.IsUnknown() {
			b.WriteString("\n")
			b.WriteString(buildWithName(name))
		}
	}

	b.WriteString("}\n\n")
	return b.String()
}

// buildWithLabel builds a withLabel directive
func buildWithLabel(label WithLabelModel) string {
	var b strings.Builder

	labelName := label.Name.ValueString()
	b.WriteString(fmt.Sprintf("  withLabel: %s {\n", labelName))

	if !label.Cpus.IsNull() && !label.Cpus.IsUnknown() {
		b.WriteString("  " + writeNumberField("cpus", label.Cpus.ValueInt64()))
	}
	if !label.Memory.IsNull() && !label.Memory.IsUnknown() {
		b.WriteString("  " + writeStringField("memory", label.Memory.ValueString()))
	}
	if !label.Time.IsNull() && !label.Time.IsUnknown() {
		b.WriteString("  " + writeStringField("time", label.Time.ValueString()))
	}
	if !label.Disk.IsNull() && !label.Disk.IsUnknown() {
		b.WriteString("  " + writeStringField("disk", label.Disk.ValueString()))
	}
	if !label.Queue.IsNull() && !label.Queue.IsUnknown() {
		b.WriteString("  " + writeStringField("queue", label.Queue.ValueString()))
	}
	if !label.ClusterOptions.IsNull() && !label.ClusterOptions.IsUnknown() {
		b.WriteString("  " + writeStringField("cluster_options", label.ClusterOptions.ValueString()))
	}
	if !label.Container.IsNull() && !label.Container.IsUnknown() {
		b.WriteString("  " + writeStringField("container", label.Container.ValueString()))
	}
	if !label.ContainerOptions.IsNull() && !label.ContainerOptions.IsUnknown() {
		b.WriteString("  " + writeStringField("container_options", label.ContainerOptions.ValueString()))
	}
	if !label.ErrorStrategy.IsNull() && !label.ErrorStrategy.IsUnknown() {
		b.WriteString("  " + writeStringField("error_strategy", label.ErrorStrategy.ValueString()))
	}
	if !label.MaxRetries.IsNull() && !label.MaxRetries.IsUnknown() {
		b.WriteString("  " + writeNumberField("max_retries", label.MaxRetries.ValueInt64()))
	}

	b.WriteString("  }\n")
	return b.String()
}

// buildWithName builds a withName directive
func buildWithName(name WithNameModel) string {
	var b strings.Builder

	pattern := name.Pattern.ValueString()
	// Pattern needs to be quoted if it contains wildcards
	b.WriteString(fmt.Sprintf("  withName: %s {\n", quoteGroovyString(pattern)))

	if !name.Cpus.IsNull() && !name.Cpus.IsUnknown() {
		b.WriteString("  " + writeNumberField("cpus", name.Cpus.ValueInt64()))
	}
	if !name.Memory.IsNull() && !name.Memory.IsUnknown() {
		b.WriteString("  " + writeStringField("memory", name.Memory.ValueString()))
	}
	if !name.Time.IsNull() && !name.Time.IsUnknown() {
		b.WriteString("  " + writeStringField("time", name.Time.ValueString()))
	}
	if !name.Disk.IsNull() && !name.Disk.IsUnknown() {
		b.WriteString("  " + writeStringField("disk", name.Disk.ValueString()))
	}
	if !name.Queue.IsNull() && !name.Queue.IsUnknown() {
		b.WriteString("  " + writeStringField("queue", name.Queue.ValueString()))
	}
	if !name.ClusterOptions.IsNull() && !name.ClusterOptions.IsUnknown() {
		b.WriteString("  " + writeStringField("cluster_options", name.ClusterOptions.ValueString()))
	}
	if !name.Container.IsNull() && !name.Container.IsUnknown() {
		b.WriteString("  " + writeStringField("container", name.Container.ValueString()))
	}
	if !name.ContainerOptions.IsNull() && !name.ContainerOptions.IsUnknown() {
		b.WriteString("  " + writeStringField("container_options", name.ContainerOptions.ValueString()))
	}
	if !name.ErrorStrategy.IsNull() && !name.ErrorStrategy.IsUnknown() {
		b.WriteString("  " + writeStringField("error_strategy", name.ErrorStrategy.ValueString()))
	}
	if !name.MaxRetries.IsNull() && !name.MaxRetries.IsUnknown() {
		b.WriteString("  " + writeNumberField("max_retries", name.MaxRetries.ValueInt64()))
	}

	b.WriteString("  }\n")
	return b.String()
}

// isProcessEmpty checks if the process block has any non-null values
func isProcessEmpty(process ProcessModel) bool {
	return (process.Executor.IsNull() || process.Executor.IsUnknown()) &&
		(process.Queue.IsNull() || process.Queue.IsUnknown()) &&
		(process.Cpus.IsNull() || process.Cpus.IsUnknown()) &&
		(process.Memory.IsNull() || process.Memory.IsUnknown()) &&
		(process.Time.IsNull() || process.Time.IsUnknown()) &&
		(process.Disk.IsNull() || process.Disk.IsUnknown()) &&
		(process.ErrorStrategy.IsNull() || process.ErrorStrategy.IsUnknown()) &&
		(process.MaxRetries.IsNull() || process.MaxRetries.IsUnknown()) &&
		(process.MaxErrors.IsNull() || process.MaxErrors.IsUnknown()) &&
		(process.ClusterOptions.IsNull() || process.ClusterOptions.IsUnknown()) &&
		(process.Container.IsNull() || process.Container.IsUnknown()) &&
		(process.ContainerOptions.IsNull() || process.ContainerOptions.IsUnknown()) &&
		len(process.WithLabel) == 0 &&
		len(process.WithName) == 0
}
