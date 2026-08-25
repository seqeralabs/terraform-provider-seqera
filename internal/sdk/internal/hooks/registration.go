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

	// Register compute environment missing hook to treat Platform's "unknown compute
	// environment" 400 as a 404 on CE describe/delete. Lets entity-missing-codes stay
	// [404], so a rejected delete is no longer silently reported as success (#240).
	computeEnvMissingHook := &ComputeEnvMissingHook{}
	h.registerAfterSuccessHook(computeEnvMissingHook)

	// Register compute environment status polling hook to wait for AVAILABLE status
	computeEnvStatusHook := &ComputeEnvStatusHook{}
	h.registerAfterSuccessHook(computeEnvStatusHook)

	// Register token list error hook to handle permission errors (401/403)
	// This allows token creation to succeed even when user can't list all tokens
	tokenListErrorHook := &TokenListErrorHook{}
	h.registerAfterSuccessHook(tokenListErrorHook)

	// exampleHook := &ExampleHook{}

	// h.registerSDKInitHook(exampleHook)
	// h.registerBeforeRequestHook(exampleHook)
	// h.registerAfterErrorHook(exampleHook)
	// h.registerAfterSuccessHook(exampleHook)
}
