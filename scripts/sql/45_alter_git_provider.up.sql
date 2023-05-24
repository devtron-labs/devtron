UPDATE git_provider
SET git_host_id=1
WHERE id = 1
  and git_host_id IS NULL;