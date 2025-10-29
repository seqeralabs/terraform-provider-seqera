package nextflow

import (
	"strings"
)

// buildSingularityBlock builds the Singularity container engine configuration
func buildSingularityBlock(singularity SingularityModel) string {
	if isSingularityEmpty(singularity) {
		return ""
	}

	var b strings.Builder
	b.WriteString("singularity {\n")

	if !singularity.Enabled.IsNull() && !singularity.Enabled.IsUnknown() {
		b.WriteString(writeBoolField("enabled", singularity.Enabled.ValueBool()))
	}
	if !singularity.AutoMounts.IsNull() && !singularity.AutoMounts.IsUnknown() {
		b.WriteString(writeBoolField("auto_mounts", singularity.AutoMounts.ValueBool()))
	}
	if !singularity.CacheDir.IsNull() && !singularity.CacheDir.IsUnknown() {
		b.WriteString(writeStringField("cache_dir", singularity.CacheDir.ValueString()))
	}
	if !singularity.RunOptions.IsNull() && !singularity.RunOptions.IsUnknown() {
		b.WriteString(writeStringField("run_options", singularity.RunOptions.ValueString()))
	}
	if !singularity.EnvWhitelist.IsNull() && !singularity.EnvWhitelist.IsUnknown() {
		b.WriteString(writeStringField("env_whitelist", singularity.EnvWhitelist.ValueString()))
	}
	if !singularity.PullTimeout.IsNull() && !singularity.PullTimeout.IsUnknown() {
		b.WriteString(writeStringField("pull_timeout", singularity.PullTimeout.ValueString()))
	}

	b.WriteString("}\n\n")
	return b.String()
}

// buildDockerBlock builds the Docker container engine configuration
func buildDockerBlock(docker DockerModel) string {
	if isDockerEmpty(docker) {
		return ""
	}

	var b strings.Builder
	b.WriteString("docker {\n")

	if !docker.Enabled.IsNull() && !docker.Enabled.IsUnknown() {
		b.WriteString(writeBoolField("enabled", docker.Enabled.ValueBool()))
	}
	if !docker.Registry.IsNull() && !docker.Registry.IsUnknown() {
		b.WriteString(writeStringField("registry", docker.Registry.ValueString()))
	}
	if !docker.RunOptions.IsNull() && !docker.RunOptions.IsUnknown() {
		b.WriteString(writeStringField("run_options", docker.RunOptions.ValueString()))
	}
	if !docker.Temp.IsNull() && !docker.Temp.IsUnknown() {
		b.WriteString(writeStringField("temp", docker.Temp.ValueString()))
	}
	if !docker.Remove.IsNull() && !docker.Remove.IsUnknown() {
		b.WriteString(writeBoolField("remove", docker.Remove.ValueBool()))
	}

	b.WriteString("}\n\n")
	return b.String()
}

// isSingularityEmpty checks if the Singularity block has any non-null values
func isSingularityEmpty(singularity SingularityModel) bool {
	return (singularity.Enabled.IsNull() || singularity.Enabled.IsUnknown()) &&
		(singularity.AutoMounts.IsNull() || singularity.AutoMounts.IsUnknown()) &&
		(singularity.CacheDir.IsNull() || singularity.CacheDir.IsUnknown()) &&
		(singularity.RunOptions.IsNull() || singularity.RunOptions.IsUnknown()) &&
		(singularity.EnvWhitelist.IsNull() || singularity.EnvWhitelist.IsUnknown()) &&
		(singularity.PullTimeout.IsNull() || singularity.PullTimeout.IsUnknown())
}

// isDockerEmpty checks if the Docker block has any non-null values
func isDockerEmpty(docker DockerModel) bool {
	return (docker.Enabled.IsNull() || docker.Enabled.IsUnknown()) &&
		(docker.Registry.IsNull() || docker.Registry.IsUnknown()) &&
		(docker.RunOptions.IsNull() || docker.RunOptions.IsUnknown()) &&
		(docker.Temp.IsNull() || docker.Temp.IsUnknown()) &&
		(docker.Remove.IsNull() || docker.Remove.IsUnknown())
}
