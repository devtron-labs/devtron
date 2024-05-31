/*
 * Copyright (c) 2024. Devtron Inc.
 */

UPDATE public.scan_tool_step
SET cli_command = 'trivy image -f json -o {{.OUTPUT_FILE_PATH}} --timeout {{.timeout}} {{.IMAGE_NAME}} --username {{.USERNAME}} --password {{.PASSWORD}} {{.EXTRA_ARGS}}'
WHERE id = 1;
