#!/usr/bin/env python3
"""
Add 409 Conflict responses to all create operations in overlay files.
"""

import yaml
import sys
from pathlib import Path

def add_409_response(operations_dict):
    """Add 409 response to an operation's responses if it doesn't exist."""
    if not isinstance(operations_dict, dict):
        return False

    modified = False

    # Look for POST operations (create operations)
    for path, methods in operations_dict.items():
        if not isinstance(methods, dict):
            continue

        # Check if this has a POST method (create operation)
        if 'post' in methods:
            post_op = methods['post']

            # Check if it's a create operation
            entity_op = post_op.get('x-speakeasy-entity-operation', '')
            if '#create' in entity_op:
                # Get or create responses
                if 'responses' not in post_op:
                    continue

                responses = post_op['responses']

                # Check if 409 already exists
                if '409' not in responses:
                    # Add 409 response
                    responses['409'] = {
                        'description': 'Conflict - resource already exists',
                        'content': {
                            'application/json': {
                                'schema': {
                                    '$ref': '#/components/schemas/ErrorResponse'
                                }
                            }
                        }
                    }
                    modified = True
                    print(f"  Added 409 to {entity_op}")

    return modified

def process_overlay_file(filepath):
    """Process a single overlay file."""
    print(f"\nProcessing {filepath.name}...")

    try:
        with open(filepath, 'r') as f:
            data = yaml.safe_load(f)

        if not data or 'actions' not in data:
            print(f"  No actions found, skipping")
            return False

        modified = False

        # Process each action
        for action in data['actions']:
            if 'update' in action:
                update_dict = action['update']
                if add_409_response(update_dict):
                    modified = True

        if modified:
            # Write back the file
            with open(filepath, 'w') as f:
                yaml.dump(data, f, default_flow_style=False, sort_keys=False, width=120)
            print(f"  ✓ Modified {filepath.name}")
            return True
        else:
            print(f"  No changes needed")
            return False

    except Exception as e:
        print(f"  ✗ Error processing {filepath.name}: {e}")
        return False

def main():
    overlays_dir = Path(__file__).parent / 'overlays'

    if not overlays_dir.exists():
        print(f"Error: overlays directory not found at {overlays_dir}")
        sys.exit(1)

    # Get all YAML files in overlays directory
    overlay_files = sorted(overlays_dir.glob('*.yaml'))

    # Skip certain files
    skip_files = {'api-description.yaml', 'schema-fixes.yaml', 'speakeasy.yaml',
                  'example-resource.yaml', 'required.yaml', 'users.yaml'}

    overlay_files = [f for f in overlay_files if f.name not in skip_files]

    print(f"Found {len(overlay_files)} overlay files to process")

    modified_count = 0
    for filepath in overlay_files:
        if process_overlay_file(filepath):
            modified_count += 1

    print(f"\n{'='*60}")
    print(f"Summary: Modified {modified_count} overlay file(s)")
    print(f"{'='*60}")

if __name__ == '__main__':
    main()
