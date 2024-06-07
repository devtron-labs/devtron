/*
 * Copyright (c) 2024. Devtron Inc.
 */

DELETE FROM plugin_step_variable WHERE name = 'JiraUsername';
DELETE FROM plugin_step_variable WHERE name = 'JiraPassword';
DELETE FROM plugin_step_variable WHERE name = 'JiraBaseUrl';
DELETE FROM plugin_step_variable WHERE name = 'JiraId';