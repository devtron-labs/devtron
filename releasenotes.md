====CRITICAL UPGRADE NOTICE - POSTGRES 11 TO 14====

IMPORTANT: This upgrade requires planned downtime
Before proceeding with the upgrade to version 1.5.0, please:

1. Review the [Upgrade Guide](https://docs.devtron.ai/upgrade/devtron-upgrade-1.5.0) for all prerequisites.

2. Deploy the [Devtron-Backup Chart](https://docs.devtron.ai/install/devtron-backup).

3. Verify at least one successful backup has been completed.

WARNING: Your Devtron will be unavailable during the entire prerequisite process until successful completion.

For questions or assistance, our team is available on [Discord](https://discord.devtron.ai/).

## v1.5.0

## Enhancements
- feat: Config Approval bypass v1 (for cm, cs and deployment template) (#6493)
- feat: Licensing changes (#6474)
## Bugs
- fix: Added support for IAM role S3 blob client (#6521)
- fix: Making tx independent of git sensor call (#6519)
- fix: Sql query memory issue (#6516)
- fix: Grafana data-source get API (#6503)
- fix: Made triggeredBy for CI & CD uniform (#6489)
- fix: HPA permission denied error (#6485)
- fix: Force abort for pre/post cd  (#6475)
- fix: Optimisation in workflow status api (#6473)
- fix: Let user delete the container reg if the ref app was deleted (#6464)
## Others
- misc: Update the descriptions of env variables (#6499)
- misc: Patch api for user attribute (#6490)
- chore: Nil implementations fixed for user attribute in telemetry (#6497)
