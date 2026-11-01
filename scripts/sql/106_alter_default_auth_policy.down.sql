UPDATE default_auth_policy set policy='{
    "data": [
        {
            "type": "p",
            "sub": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "applications",
            "act": "*",
            "obj": "{{.TeamObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "environment",
            "act": "*",
            "obj": "{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "team",
            "act": "*",
            "obj": "{{.TeamObj}}"
        },
        {
            "type": "p",
            "sub": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "user",
            "act": "*",
            "obj": "{{.TeamObj}}"
        },
        {
            "type": "p",
            "sub": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "notification",
            "act": "*",
            "obj": "{{.TeamObj}}"
        },
        {
            "type": "p",
            "sub": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "global-environment",
            "act": "*",
            "obj": "{{.EnvObj}}"
        }
    ]
}' where role_type='manager';

UPDATE default_auth_policy set policy='{
    "data": [
        {
            "type": "p",
            "sub": "role:admin_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "applications",
            "act": "*",
            "obj": "{{.TeamObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:admin_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "environment",
            "act": "*",
            "obj": "{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:admin_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "team",
            "act": "get",
            "obj": "{{.TeamObj}}"
        },
        {
            "type": "p",
            "sub": "role:admin_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "global-environment",
            "act": "get",
            "obj": "{{.EnvObj}}"
        }
    ]
}' where role_type='admin';

UPDATE default_auth_policy set policy='{
    "data": [
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "applications",
            "act": "get",
            "obj": "{{.TeamObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "applications",
            "act": "trigger",
            "obj": "{{.TeamObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "environment",
            "act": "trigger",
            "obj": "{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "environment",
            "act": "get",
            "obj": "{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "global-environment",
            "act": "get",
            "obj": "{{.EnvObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "team",
            "act": "get",
            "obj": "{{.TeamObj}}"
        }
    ]
}' where role_type='trigger';


UPDATE default_auth_policy set policy='{
    "data": [
        {
            "type": "p",
            "sub": "role:view_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "applications",
            "act": "get",
            "obj": "{{.TeamObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:view_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "environment",
            "act": "get",
            "obj": "{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:view_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "global-environment",
            "act": "get",
            "obj": "{{.EnvObj}}"
        },
        {
            "type": "p",
            "sub": "role:view_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "team",
            "act": "get",
            "obj": "{{.TeamObj}}"
        }
    ]
}' where role_type='view';
