UPDATE scan_tool_metadata
SET result_descriptor_template = '[{{$size1:= len .Results}}{{range $i1, $v1 := .Results}}{{ if  $v1.Vulnerabilities}}{{$size2:= len $v1.Vulnerabilities}}{{range $i2, $v2 := $v1.Vulnerabilities}}{{if and (eq $i1 (add $size1 -1)) (eq $i2 (add $size2 -1)) }}
{
"package": "{{$v2.PkgName}}",
"packageVersion": "{{$v2.InstalledVersion}}",
"fixedInVersion": "{{$v2.FixedVersion}}",
"severity": "{{$v2.Severity}}",
"name": "{{$v2.VulnerabilityID}}"
}{{else}}{
"package": "{{$v2.PkgName}}",
"packageVersion": "{{$v2.InstalledVersion}}",
"fixedInVersion": "{{$v2.FixedVersion}}",
"severity": "{{$v2.Severity}}",
"name": "{{$v2.VulnerabilityID}}"
},{{end}}{{end}}{{end}}{{end}}]' where name = 'TRIVY' and version ='V1';
