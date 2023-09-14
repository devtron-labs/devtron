## v0.6.22



## v0.6.22

## Bugs
- fix: updated adapter for cluster object (#3900)
- fix: rbac-modification for cluster list (#3767)
- fix: Helm app deployment history page breaking due to user details not found (#3873)
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
- feat: added new statefulset-5-0-0 chart in reference chart  (#3909)
- feat: added configurable provenance flag for buildx builds (#3905)
- feat: deployment history release not found err handling (#3811)
- feat: added validation for create app workflow API (#3842)
- feat: custom chart download (#3801)
- feat: Virtual cluster v3 (#3764)
- feat: Maintaining audit logs (#3763)
- feat: Capability to block deployments in case of vulnerabilities only if FIXED IN VERSION available  (#3796)
## Documentation
- doc: draft version of Graviton benchmark (#3890)
- doc: Okta SSO Configuration Doc (#3876)
## Others
- chore: changes for  migration no conflict (#3919)
- chore: Changed in Docker file for SQL file (#3904)
- chore: adjust duplicate action threshold (#3879)
- chore: find potential-duplicate issues  (#3858)
- chore: Update pr-issue-validator.yaml (#3849)


