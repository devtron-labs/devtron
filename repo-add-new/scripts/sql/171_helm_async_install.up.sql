
alter table installed_app_version_history
    add column helm_release_status_config text;
-- helm_release_status_config - {InstallAppVersionHistoryId: "", ReleaseInstalled: "true/false", Message: "failure reason"  }

