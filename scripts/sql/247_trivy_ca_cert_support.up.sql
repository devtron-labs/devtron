UPDATE public.scan_tool_step
SET cli_command = '{{if .CA_CERT_FILE_PATH}} SSL_CERT_FILE={{.CA_CERT_FILE_PATH}} {{end}} trivy {{if .insecure}} --insecure {{end}} image -f json -o {{.OUTPUT_FILE_PATH}} --timeout {{.timeout}} {{
.IMAGE_NAME}} --username {{.USERNAME}} --password {{.PASSWORD}}'
WHERE id = 1;
