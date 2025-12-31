## v2.0.0

## Enhancements
- feat: Rollout 5.2.0 (#6889)
- feat: Added support for tcp in virtual service and changed the apiVersion for externalSecrets (#6892)
- feat: add helm_take_ownership and helm_redeployment_request columns to user_deployment_request table (#6888)
- feat: Revamped Devtron UI with multiple dashboards (#6884)
- feat: Added support to override container name (#6880)
- feat: Increase max length for TeamRequest name field (#6876)
- feat: Added namespace support for virtualService and destinationRule (#6868)
- feat: feature flag for encryption (#6856)
- feat: encryption for db credentials (#6852)
## Bugs
- fix: migrate proxy chart dependencies and refactor related functions (#6899)
- fix: enhance validation and error handling in cluster update process (#6887)
- fix: Invalid type casting error for custom charts (#6883)
- fix: validation on team name (#6872)
- fix: sql injection  (#6861)
- fix: user manager fix (#6854)
## Others
- misc: Add support for migrating plugin metadata to parent metadata (#6902)
- misc: update UserDeploymentRequestWithAdditionalFields struct to include tableName for PostgreSQL compatibility (#6896)
- chore: rename SQL migration files for consistency (#6885)
- misc: Vc empty ns fix (#6871)
- misc: added validation on create environment (#6859)
- misc: migration unique constraint on mpc (#6851)
- misc: helm app details API spec (#6850)
- misc: api Spec Added for draft (#6849)
- misc: api Specs added for lock config (#6847)


