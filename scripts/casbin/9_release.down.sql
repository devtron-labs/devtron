/*
 * Copyright (c) 2024. Devtron Inc.
 */

DELETE FROM casbin_rule where v0='role:super-admin___' and v1='release';
DELETE FROM casbin_rule where v0='role:super-admin___' and v1='release-requirement';
DELETE FROM casbin_rule where v0='role:super-admin___' and v1='release-track';
DELETE FROM casbin_rule where v0='role:super-admin___' and v1='release-track-requirement';