package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ConflictErrorHook handles 409 Conflict errors by providing clear error messages
// indicating that a resource already exists. This prevents unnecessary retries
// and helps users understand they need to import or use a different name.
type ConflictErrorHook struct{}

// AfterError implements the afterErrorHook interface
func (h *ConflictErrorHook) AfterError(hookCtx AfterErrorContext, res *http.Response, err error) (*http.Response, error) {
	// Only process if we have a response and it's a 409
	if res == nil || res.StatusCode != 409 {
		return res, err
	}

	// Check if this is a create operation
	// All create operations follow the pattern "Create*"
	if !strings.HasPrefix(hookCtx.OperationID, "Create") {
		return res, err
	}

	// Read the response body to extract error details
	bodyBytes, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		// If we can't read the body, return the original error
		return res, err
	}
	res.Body.Close()

	// Try to parse the error response to get more details
	var errorResponse struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(bodyBytes, &errorResponse)

	// Determine resource type and extract resource name from the API message
	var resourceType, resourceName, workspace string

	switch {
	case strings.Contains(hookCtx.OperationID, "ComputeEnv"):
		resourceType = "Compute environment"
		// Extract name from message like "A compute environment with name 'aws-batch-spot-4' already exists within the workspace 'genomics-research'"
		if strings.Contains(errorResponse.Message, "compute environment with name") {
			if start := strings.Index(errorResponse.Message, "'"); start != -1 {
				if end := strings.Index(errorResponse.Message[start+1:], "'"); end != -1 {
					resourceName = errorResponse.Message[start+1 : start+1+end]
				}
			}
			if strings.Contains(errorResponse.Message, "within the workspace") {
				if start := strings.LastIndex(errorResponse.Message, "'"); start != -1 {
					if strings.Count(errorResponse.Message, "'") >= 4 {
						// Find second-to-last quote
						secondLastQuote := strings.LastIndex(errorResponse.Message[:start], "'")
						if secondLastQuote != -1 {
							workspace = errorResponse.Message[secondLastQuote+1 : start]
						}
					}
				}
			}
		}
	case strings.Contains(hookCtx.OperationID, "Credential"):
		resourceType = "Credential"
	case strings.Contains(hookCtx.OperationID, "Pipeline"):
		resourceType = "Pipeline"
	case strings.Contains(hookCtx.OperationID, "Action"):
		resourceType = "Action"
	case strings.Contains(hookCtx.OperationID, "Studio"):
		resourceType = "Data studio"
	case strings.Contains(hookCtx.OperationID, "Workspace"):
		resourceType = "Workspace"
	default:
		resourceType = "Resource"
	}

	// Create a helpful error message
	var helpfulMessage string

	if resourceName != "" {
		if workspace != "" {
			helpfulMessage = fmt.Sprintf(
				"%s '%s' already exists in workspace '%s'.\n\n"+
					"To resolve:\n"+
					"  - Use 'terraform import' to import the existing resource\n"+
					"  - Use a different name for this resource\n"+
					"  - Delete the existing resource from Seqera Platform first",
				resourceType, resourceName, workspace,
			)
		} else {
			helpfulMessage = fmt.Sprintf(
				"%s '%s' already exists.\n\n"+
					"To resolve:\n"+
					"  - Use 'terraform import' to import the existing resource\n"+
					"  - Use a different name for this resource\n"+
					"  - Delete the existing resource from Seqera Platform first",
				resourceType, resourceName,
			)
		}
	} else {
		// Fallback if we couldn't parse the name
		helpfulMessage = fmt.Sprintf(
			"%s already exists. This may happen if:\n"+
				"  - A resource with this name already exists in the workspace\n"+
				"  - The resource was created outside Terraform\n"+
				"  - The Terraform state is out of sync\n\n"+
				"To resolve:\n"+
				"  - Use 'terraform import' to import the existing resource\n"+
				"  - Use a different name for this resource\n"+
				"  - Delete the existing resource from Seqera Platform first",
			resourceType,
		)
	}

	// Return nil response and the helpful error message
	// This will be displayed by Terraform without the "failure to invoke API" wrapper
	return nil, fmt.Errorf(helpfulMessage)
}
