package hooks

/*
 * This file is only ever generated once on the first generation and then is free to be modified.
 * Any hooks you wish to add should be registered in the initHooks function. Feel free to define
 * your hooks in this file or in separate files in the hooks package.
 *
 * Hooks are registered per SDK instance, and are valid for the lifetime of the SDK instance.
 */

func initHooks(h *Hooks) {
	// Register generic resource error hook to treat 403 as 404 for all deleted resources
	// This handles all describe operations uniformly across all resource types
	genericResourceErrorHook := &GenericResourceErrorHook{}
	h.registerAfterSuccessHook(genericResourceErrorHook)

	// Register compute environment status polling hook to wait for AVAILABLE status
	computeEnvStatusHook := &ComputeEnvStatusHook{}
	h.registerAfterSuccessHook(computeEnvStatusHook)

	// Register token list error hook to handle permission errors (401/403)
	// This allows token creation to succeed even when user can't list all tokens
	tokenListErrorHook := &TokenListErrorHook{}
	h.registerAfterSuccessHook(tokenListErrorHook)

	// Register agent binding conflict hook so a service account binding error
	// on agent create shows the Platform's message instead of "already exists"
	agentBindingConflictHook := &AgentBindingConflictHook{}
	h.registerAfterSuccessHook(agentBindingConflictHook)

	// Register lineage settings status hook to wait for AUTO provisioning to finish
	// (status PENDING -> AVAILABLE/ERRORED) after the settings update
	lineageSettingsStatusHook := &LineageSettingsStatusHook{}
	h.registerAfterSuccessHook(lineageSettingsStatusHook)

	// Register lineage settings not-configured hook to turn the 200-with-empty-fields
	// response for a workspace without lineage into a 404 so Terraform drops the resource
	lineageSettingsNotConfiguredHook := &LineageSettingsNotConfiguredHook{}
	h.registerAfterSuccessHook(lineageSettingsNotConfiguredHook)

	// Register compute environment delete error hook so a 400/409 rejection of a
	// deletion fails the destroy with the Platform's message instead of being
	// treated as success
	ceDeleteErrorHook := &CEDeleteErrorHook{}
	h.registerAfterSuccessHook(ceDeleteErrorHook)

	// exampleHook := &ExampleHook{}

	// h.registerSDKInitHook(exampleHook)
	// h.registerBeforeRequestHook(exampleHook)
	// h.registerAfterErrorHook(exampleHook)
	// h.registerAfterSuccessHook(exampleHook)
}
