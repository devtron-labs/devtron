## v0.6.11

## Bugs
- fix: Nats server version upgrade (#2837)
- fix: K8s side resources error handled some unable to fetch some kinds in cluster (#2896)
- fix: helm and gitOps pipeline feature gap changes (#2891)
- fix: Adding /Deleting git material reflection in workflow editor and build & deploy (#2744)
- fix: issue resolved for notification on deployment trigger and telemetry summary event (#2893)
- fix: helm manifest output in deployment history (#2822)
- fix: sending appName in the response of getAppWorkflow API (#2879)
- fix: cluster not reachable handling (#2872)
## Enhancements
- feat: Support of reuse container image in containers section in reference chart 4-17 (#2675)
- feat: Added Rollout chart 4.17 and Deployment chart 1.1 with canary using flagger and Istio support (#2910)
- feat: Codacy plugin integration (#2892)
- feat: Semgrep plugin integration (#2877)
- perf: Tracing observability tool integration (#2782)
## Documentation
- docs: removed command to install a particular version of Devtron (#2888)
- docs: Updated README.md file for Devtron chart (#2887)
- docs: added manage tags (#2907)
- docs: updated target platform (#2857)
## Others
- task: added new fields for eso secrets (#2899)
- task: updated deployment status timeline api for helm based pipelines (#2897)
- task: updated logic for adding health status timeline (#2870)


## v0.6.11-rc.0

## Bugs
- fix: bitbucket gitops config fixes (#2768)
- fix: updated sonarqube plugin script (#2855)
- fix: updated the description in sonarqube migration (#2853)
- fix: External Ci Trigger - Won't show images for manual Cd pipelines (#2802)
- fix: app create api bug fix if chart ref is different (#2851)
- fix: Dockerfile git repo delete tooltip view to build config (#2792)
## Enhancements
- feat: Option for changing project in Helm apps and project assignment to CLI helm apps. (#2780)
- feat: Audit logs for resource deletion from devtron UI  (#2791)
- feat: verify token function scope made public (#2856)
- feat: Sonarqube polling script update (#2846)
- feat: deployment status restructuring (#2825)
## Documentation
- docs: updated aws external secret doc (#2781)
- docs: updated broken links and minor updates (resource browser) (#2844)
- docs: added cluster, k8s resource browser, k8s resource permission (#2787)
