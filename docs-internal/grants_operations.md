### Compute

| Permission                     | Description                                    | API endpoint                                |
| ------------------------------ | ---------------------------------------------- | ------------------------------------------- |
| **compute_environment:read**   | List all compute environments                  | `GET /compute-envs`                         |
|                                | View compute environment details               | `GET /compute-envs/{computeEnvId}`          |
| **compute_environment:write**  | Create a new compute environment               | `POST /compute-envs`                        |
|                                | Edit an existing compute environment           | `PUT /compute-envs/{computeEnvId}`          |
|                                | Set a compute environment as primary           | `POST /compute-envs/{computeEnvId}/primary` |
|                                | Disable compute environment                    | `POST /compute-envs/{computeEnvId}/disable` |
|                                | Enable compute environment                     | `POST /compute-envs/{computeEnvId}/enable`  |
|                                | Validate compute environment name availability | `GET /compute-envs/validate`                |
| **compute_environment:delete** | Delete a compute environment                   | `DELETE /compute-envs/{computeEnvId}`       |
| **credentials:read**           | List all credentials in workspace              | _(Used by Platform)_                        |
|                                | View credential details                        | _(Used by Platform)_                        |
| **credentials:write**          | Add new credentials                            | _(Used by Platform)_                        |
|                                | Edit existing credentials                      | _(Used by Platform)_                        |
|                                | Validate credentials                           | _(Used by Platform)_                        |
|                                | Validate credential name availability          | _(Used by Platform)_                        |
| **credentials:delete**         | Delete credentials                             | _(Used by Platform)_                        |
| **credentials_encrypted:read** | Get encrypted credentials                      | _(Used by Platform)_                        |
| **pipeline_secrets:read**      | List all pipeline secrets                      | `GET /pipeline-secrets`                     |
|                                | View pipeline secret details                   | `GET /pipeline-secrets/{secretId}`          |
| **pipeline_secrets:write**     | Create a new pipeline secret                   | `POST /pipeline-secrets`                    |
|                                | Validate secret name availability              | `GET /pipeline-secrets/validate`            |
|                                | Edit an existing pipeline secret               | `PUT /pipeline-secrets/{secretId}`          |
| **pipeline_secrets:delete**    | Delete a pipeline secret                       | `DELETE /pipeline-secrets/{secretId}`       |
| **platform:read**              | List available platforms                       | _(Used by Platform)_                        |
|                                | List platform regions                          | _(Used by Platform)_                        |
|                                | View platform details                          | _(Used by Platform)_                        |

### Data

| Permission              | Description                                         | API endpoint                   |
| ----------------------- | --------------------------------------------------- | ------------------------------ |
| **data_link:read**      | List all data-links (cloud buckets)                 | _(Used by Platform)_           |
|                         | Browse data-link contents                           | _(Used by Platform)_           |
|                         | Browse data-link contents at the given path         | _(Used by Platform)_           |
|                         | View data-link details                              | _(Used by Platform)_           |
| **data_link:write**     | Refresh data-link cache                             | _(Used by Platform)_           |
|                         | Browse data-link directory tree                     | _(Used by Platform)_           |
|                         | Download files from data-link                       | _(Used by Platform)_           |
|                         | Generate download URL for data-link files           | _(Used by Platform)_           |
|                         | Generate download script                            | _(Used by Platform)_           |
|                         | Upload files to data-link                           | _(Used by Platform)_           |
|                         | Upload files to data-link at the given path         | _(Used by Platform)_           |
|                         | Complete file upload to data-link                   | _(Used by Platform)_           |
|                         | Complete file upload to data-link at the given path | _(Used by Platform)_           |
|                         | Create a custom data-link                           | _(Used by Platform)_           |
|                         | Edit data-link metadata                             | _(Used by Platform)_           |
| **data_link:delete**    | Delete files from data-link                         | _(Used by Platform)_           |
|                         | Remove a data-link from workspace                   | _(Used by Platform)_           |
| **data_link:admin**     | Hide data-links                                     | _(Used by Platform)_           |
|                         | Show data-links                                     | _(Used by Platform)_           |
| **dataset:read**        | List datasets                                       | _(Used by Platform)_           |
|                         | List workspace dataset versions                     | _(Used by Platform)_           |
|                         | List dataset versions                               | _(Used by Platform)_           |
|                         | View dataset metadata                               | _(Used by Platform)_           |
|                         | Download dataset                                    | _(Used by Platform)_           |
|                         | List all datasets                                   | _(Used by Platform)_           |
|                         | List latest dataset versions                        | _(Used by Platform)_           |
|                         | List versions for a specific dataset                | _(Used by Platform)_           |
|                         | List datasets used in a pipeline launch             | _(Used by Platform)_           |
|                         | View dataset metadata                               | _(Used by Platform)_           |
|                         | Download dataset files                              | _(Used by Platform)_           |
| **dataset:write**       | Create dataset                                      | _(Used by Platform)_           |
|                         | Edit dataset                                        | _(Used by Platform)_           |
|                         | Upload dataset                                      | _(Used by Platform)_           |
|                         | Create a new dataset                                | _(Used by Platform)_           |
|                         | Edit dataset metadata                               | _(Used by Platform)_           |
|                         | Upload files to dataset                             | _(Used by Platform)_           |
| **dataset:delete**      | Delete dataset                                      | _(Used by Platform)_           |
|                         | Delete a single dataset                             | _(Used by Platform)_           |
|                         | Delete multiple datasets                            | _(Used by Platform)_           |
| **dataset:admin**       | Hide any workspace user's datasets                  | _(Used by Platform)_           |
|                         | Show any workspace user's datasets                  | _(Used by Platform)_           |
|                         | Disable any workspace user's dataset version        | _(Used by Platform)_           |
| **dataset_label:write** | Add labels to datasets                              | `POST /datasets/labels/add`    |
|                         | Remove labels from datasets                         | `POST /datasets/labels/remove` |
|                         | Apply label sets to datasets                        | `POST /datasets/labels/apply`  |

### Pipelines

| Permission                 | Description                                                     | API endpoint                    |
| -------------------------- | --------------------------------------------------------------- | ------------------------------- |
| **action:read**            | View action details                                             | _(Used by Platform)_            |
|                            | View available action types                                     | _(Used by Platform)_            |
|                            | List all actions in workspace                                   | _(Used by Platform)_            |
| **action:execute**         | Trigger an action to run                                        | _(Used by Platform)_            |
| **action:write**           | Create a new action                                             | _(Used by Platform)_            |
|                            | Edit an existing action                                         | _(Used by Platform)_            |
|                            | Test action configuration                                       | _(Used by Platform)_            |
|                            | Pause a running action                                          | _(Used by Platform)_            |
|                            | Validate action name availability                               | _(Used by Platform)_            |
| **action:delete**          | Delete an action                                                | _(Used by Platform)_            |
| **action_label:write**     | Apply resource labels when adding an action                     | _(Used by Platform)_            |
|                            | Apply resource labels when editing an action                    | _(Used by Platform)_            |
|                            | Add labels to actions                                           | `POST /actions/labels/add`      |
|                            | Remove labels from actions                                      | `POST /actions/labels/remove`   |
|                            | Apply label sets to actions                                     | `POST /actions/labels/apply`    |
| **container:read**         | View container details                                          | _(Used by Platform)_            |
|                            | List containers                                                 | _(Used by Platform)_            |
|                            | List workflow containers                                        | _(Used by Platform)_            |
| **launch:read**            | View launch details                                             | `GET /launch/{launchId}`        |
| **pipeline:read**          | View pipeline repository information                            | _(Used by Platform)_            |
|                            | View pipeline schema and parameters                             | _(Used by Platform)_            |
|                            | View pipeline schema from repository URL                        | _(Used by Platform)_            |
|                            | View pipeline launch configuration                              | _(Used by Platform)_            |
|                            | List available pipeline repositories                            | _(Used by Platform)_            |
|                            | List all pipelines in workspace                                 | _(Used by Platform)_            |
|                            | View pipeline details                                           | _(Used by Platform)_            |
|                            | List pipeline versions                                          | _(Used by Platform)_            |
|                            | Fetch pipeline optimization                                     | _(Used by Platform)_            |
| **pipeline:write**         | Modify pipeline details when launching a pipeline run           | _(Used by Platform)_            |
|                            | Add a new pipeline to workspace                                 | _(Used by Platform)_            |
|                            | Edit pipeline (default version) configuration                   | _(Used by Platform)_            |
|                            | Configure pipeline                                              | _(Used by Platform)_            |
|                            | Validate pipeline name availability                             | _(Used by Platform)_            |
|                            | Create a pipeline schema                                        | _(Used by Platform)_            |
|                            | Validate pipeline version name availability                     | _(Used by Platform)_            |
|                            | Manage pipeline version                                         | _(Used by Platform)_            |
|                            | Edit pipeline version configuration                             | _(Used by Platform)_            |
| **pipeline:delete**        | Delete a pipeline                                               | _(Used by Platform)_            |
| **pipeline_label:write**   | Apply resource labels when launching a pipeline run             | _(Used by Platform)_            |
|                            | Add labels to pipelines                                         | `POST /pipelines/labels/add`    |
|                            | Apply resource labels when adding a pipeline                    | _(Used by Platform)_            |
|                            | Apply resource labels when editing a pipeline (default version) | _(Used by Platform)_            |
|                            | Apply resource labels when editing a pipeline version           | _(Used by Platform)_            |
|                            | Remove labels from pipelines                                    | `POST /pipelines/labels/remove` |
|                            | Apply label sets to pipelines                                   | `POST /pipelines/labels/apply`  |
| **workflow:read**          | View run details                                                | _(Used by Platform)_            |
|                            | View run progress                                               | _(Used by Platform)_            |
|                            | List tasks in a run                                             | _(Used by Platform)_            |
|                            | View individual task details                                    | _(Used by Platform)_            |
|                            | View run metrics                                                | _(Used by Platform)_            |
|                            | List all runs in workspace                                      | _(Used by Platform)_            |
|                            | View run launch configuration                                   | _(Used by Platform)_            |
|                            | View run execution logs                                         | _(Used by Platform)_            |
|                            | View task-specific logs                                         | _(Used by Platform)_            |
|                            | Download run logs                                               | _(Used by Platform)_            |
|                            | Download run content in a workspace                             | _(Used by Platform)_            |
|                            | Download task logs                                              | _(Used by Platform)_            |
|                            | View run reports                                                | _(Used by Platform)_            |
|                            | Download run report                                             | _(Used by Platform)_            |
|                            | Fetch workflow optimization                                     | _(Used by Platform)_            |
|                            | Check optimized workflow list                                   | _(Used by Platform)_            |
| **workflow:execute**       | Launch a pipeline run                                           | _(Used by Platform)_            |
|                            | Cancel a running pipeline                                       | _(Used by Platform)_            |
|                            | Launch a pipeline run                                           | _(Used by Platform)_            |
| **workflow:write**         | Create execution trace                                          | _(Used by Platform)_            |
|                            | Update trace heartbeat                                          | _(Used by Platform)_            |
|                            | Mark trace begin                                                | _(Used by Platform)_            |
|                            | Mark trace complete                                             | _(Used by Platform)_            |
|                            | Update trace progress                                           | _(Used by Platform)_            |
| **workflow:delete**        | Delete a single run                                             | _(Used by Platform)_            |
|                            | Delete multiple runs                                            | _(Used by Platform)_            |
| **workflow_label:write**   | Add labels to runs                                              | _(Used by Platform)_            |
|                            | Remove labels from runs                                         | _(Used by Platform)_            |
|                            | Apply label sets to runs                                        | _(Used by Platform)_            |
| **workflow_quick:execute** | Launch quick pipeline                                           | _(Used by Platform)_            |
|                            | Launch quick pipeline                                           | _(Used by Platform)_            |
|                            | GA4GH: create a run                                             | `POST /ga4gh/wes/v1/runs`       |
| **workflow_star:read**     | Check if run is starred (favourited)                            | _(Used by Platform)_            |
| **workflow_star:write**    | Star (favourite) a run                                          | _(Used by Platform)_            |
| **workflow_star:delete**   | Unstar (unfavourite) a run                                      | _(Used by Platform)_            |

### Settings

| Permission                 | Description                                      | API endpoint                                                                                    |
| -------------------------- | ------------------------------------------------ | ----------------------------------------------------------------------------------------------- |
| **label:read**             | List all workspace labels                        | `GET /labels`                                                                                   |
| **label:write**            | Create a new label                               | `POST /labels`                                                                                  |
|                            | Edit an existing label                           | `PUT /labels/{labelId}`                                                                         |
| **label:delete**           | Delete a label                                   | `DELETE /labels/{labelId}`                                                                      |
| **workspace:read**         | View workspace details                           | `GET /orgs/{orgId}/workspaces/{workspaceId}`                                                    |
|                            | List workspace participants                      | `GET /orgs/{orgId}/workspaces/{workspaceId}/participants`                                       |
| **workspace:write**        | Edit workspace settings                          | `PUT /orgs/{orgId}/workspaces/{workspaceId}`                                                    |
|                            | Add a workspace participant                      | `PUT /orgs/{orgId}/workspaces/{workspaceId}/participants/add`                                   |
|                            | Find workspace participant candidates            | _(Used by Platform)_                                                                            |
|                            | Change participant role                          | `PUT /orgs/{orgId}/workspaces/{workspaceId}/participants/{participantId}/role`                  |
|                            | Remove a workspace participant (user or team)    | `DELETE /orgs/{orgId}/workspaces/{workspaceId}/participants/{participantId}`                    |
|                            | Remove a workspace user (member or collaborator) | `DELETE /orgs/{orgId}/workspaces/{workspaceId}/users/{userId}`                                  |
| **workspace:delete**       | Delete the workspace                             | `DELETE /orgs/{orgId}/workspaces/{workspaceId}`                                                 |
| **workspace:admin**        | Change participant role to/from Owner            | Sub-operation on `PUT /orgs/{orgId}/workspaces/{workspaceId}/participants/{participantId}/role` |
|                            | Remove a workspace Owner by participantId        | Sub-operation on `DELETE /orgs/{orgId}/workspaces/{workspaceId}/participants/{participantId}`   |
|                            | Remove a workspace Owner by userId               | Sub-operation on `DELETE /orgs/{orgId}/workspaces/{workspaceId}/users/{userId}`                 |
| **workspace_self:delete**  | Leave workspace (remove self as participant)     | `DELETE /orgs/{orgId}/workspaces/{workspaceId}/participants`                                    |
| **workspace_studio:read**  | View studio settings for workspace               | `GET /orgs/{orgId}/workspaces/{workspaceId}/settings/studios`                                   |
| **workspace_studio:write** | Edit studio settings for workspace               | `PUT /orgs/{orgId}/workspaces/{workspaceId}/settings/studios`                                   |

### Studios

| Permission                 | Description                                                    | API endpoint         |
| -------------------------- | -------------------------------------------------------------- | -------------------- |
| **studio:read**            | View studio session details                                    | _(Used by Platform)_ |
|                            | View studio repository details                                 | _(Used by Platform)_ |
|                            | List all studios in workspace                                  | _(Used by Platform)_ |
|                            | List available studio templates                                | _(Used by Platform)_ |
|                            | List checkpoints for a studio                                  | _(Used by Platform)_ |
|                            | View checkpoint details                                        | _(Used by Platform)_ |
| **studio:execute**         | List mounted data-links for studios                            | _(Used by Platform)_ |
|                            | Start a studio session                                         | _(Used by Platform)_ |
|                            | Stop a studio session                                          | _(Used by Platform)_ |
| **studio:write**           | Create a new studio                                            | _(Used by Platform)_ |
|                            | Edit checkpoint name                                           | _(Used by Platform)_ |
|                            | Validate studio name availability                              | _(Used by Platform)_ |
| **studio:delete**          | Delete a studio                                                | _(Used by Platform)_ |
| **studio:admin**           | Delete another user's private studio                           | _(Used by Platform)_ |
|                            | Start another user's private studio                            | _(Used by Platform)_ |
|                            | Stop another user's private studio                             | _(Used by Platform)_ |
|                            | Extend another user's private studio session lifespan (iframe) | _(Used by Platform)_ |
|                            | Extend another user's private studio session lifespan          | _(Used by Platform)_ |
|                            | Administer another user's private studio                       | _(Used by Platform)_ |
| **studio_label:write**     | Apply resource labels when starting a studio                   | _(Used by Platform)_ |
| **studio_session:read**    | Open a studio                                                  | _(Used by Platform)_ |
| **studio_session:execute** | Extend studio session lifespan (iframe)                        | _(Used by Platform)_ |
|                            | Extend studio session lifespan                                 | _(Used by Platform)_ |
