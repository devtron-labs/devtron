## v2.0.0
Devtron 2.0 focuses on improving Kubernetes operability at scale. It introduces centralized platform visibility, a restructured UI for faster navigation, and foundational security improvements—reducing operational overhead across CI/CD, infrastructure, and access management.

---

## Enhancements

### Overview Dashboards
We've introduced centralized overview dashboards to give teams instant visibility across critical dimensions of the system.

- **Applications Overview**  
  View application health metrics and CI/CD pipeline activity in one place to quickly assess delivery performance and identify issues early.

- **Infrastructure Overview**  
  Gain clear insights into cluster utilization and resource allocation for better visibility into Kubernetes capacity, usage, and optimization opportunities.

- **Security Overview**  
  Get an aggregated view of your security posture to identify risks, track vulnerabilities, and monitor overall security status effectively.

---

### Reimagined Platform UI
Devtron's UI has been restructured into logical modules, making the platform more intuitive, discoverable, and easier to navigate.

- **Application Management**  
  Deploy and manage applications, access application groups, and apply bulk edits from a unified workspace.

- **Infrastructure Management**  
  Manage applications deployed via Helm, Argo CD, and Flux CD, access the chart store, and explore your Kubernetes resources using the resource browser.

- **Security Centre**  
  Review vulnerability reports, manage security scans, and enforce security policies from a single control plane.

- **Automation Enablement**  
  Configure and schedule job orchestration to power automated workflows and reduce manual operational overhead.

- **Global Configuration**  
  Configure SSO, manage clusters and environments, register container registries, and define authorization and access policies centrally.

---

### Command Bar for Faster Navigation
Use the Command Bar (Ctrl + K on Windows/Linux, Cmd + K on macOS) to quickly jump to any screen across the platform and revisit recently accessed items—reducing clicks and speeding up navigation between workflows.

---

### Additional Enhancements
- feat: Rollout 5.2.0 (#6889)
- feat: Added support for tcp in virtual service and changed the apiVersion for externalSecrets (#6892)
- feat: add helm_take_ownership and helm_redeployment_request columns to user_deployment_request table (#6888)
- feat: Added support to override container name (#6880)
- feat: Increase max length for TeamRequest name field (#6876)
- feat: Added namespace support for virtualService and destinationRule (#6868)
- feat: feature flag for encryption (#6856)
- feat: encryption for db credentials (#6852)

---

## Bugs
- fix: migrate proxy chart dependencies and refactor related functions (#6899)
- fix: enhance validation and error handling in cluster update process (#6887)
- fix: Invalid type casting error for custom charts (#6883)
- fix: validation on team name (#6872)
- fix: sql injection  (#6861)
- fix: user manager fix (#6854)

---

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
