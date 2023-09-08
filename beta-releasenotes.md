## Bugs
- fix: ci pip status query optmization (#3877)
- fix: migration script for virtual cluster v3 (#3870)
- fix: cloning app cmcs global boolean value fix (#3862)
- fix: Makefile correction (#3852)
- fix: deleting pipeline stage and related data if no stage steps are found (#3832)
- fix: Port number fix in helm app (#3843)
- fix: External cm and secret in jobs not getting added as env variable in container (#3815)
- fix: pre-cd pod not getting scheduled when node affinity is not present in external cluster. (#3806)
- fix: k8s permission and chart-group permission not getting deleted from orchestrator (#3824)
- fix: added missing audit logs while deleting cd pipeline (#3822)
## Enhancements
- feat: deployment history release not found err handling (#3811)
- feat: added validation for create app workflow API (#3842)
- feat: custom chart download (#3801)
- feat: Virtual cluster v3 (#3764)
- feat: Maintaining audit logs (#3763)
- feat: Capability to block deployments in case of vulnerabilities only if FIXED IN VERSION available  (#3796)
## Documentation
## Others
- chore: find potential-duplicate issues  (#3858)
- chore: Update pr-issue-validator.yaml (#3849)
