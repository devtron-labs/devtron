/*
 * Copyright (c) 2024. Devtron Inc.
 */

DELETE FROM plugin_step_variable where name='GcpServiceAccountJson';
DELETE FROM plugin_step_variable where name='GcpProjectName';
UPDATE plugin_step_variable SET description='Provide which cloud storage provider you want to use: "aws" for Amazon S3 or "azure" for Azure Blob Storage' WHERE name='CloudProvider';
