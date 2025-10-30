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

	// Register conflict error hook to provide clear messages for 409 errors
	// This helps users understand when a resource already exists
	conflictErrorHook := &ConflictErrorHook{}
	h.registerAfterErrorHook(conflictErrorHook)

	// Register compute environment status polling hook to wait for AVAILABLE status
	computeEnvStatusHook := &ComputeEnvStatusHook{}
	h.registerAfterSuccessHook(computeEnvStatusHook)

	// exampleHook := &ExampleHook{}

	// h.registerSDKInitHook(exampleHook)
	// h.registerBeforeRequestHook(exampleHook)
	// h.registerAfterErrorHook(exampleHook)
	// h.registerAfterSuccessHook(exampleHook)
}
