ALTER TABLE ci_workflow ADD pod_name text;
Update ci_workflow SET pod_name = name;


ALTER TABLE cd_workflow_runner ADD pod_name text;
Update cd_workflow_runner SET pod_name = name;