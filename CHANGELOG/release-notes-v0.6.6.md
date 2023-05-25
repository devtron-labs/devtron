## v0.6.6

## Bugs
- fix: Updated username for Prakarsh (#2575)
- fix: showing cicd module status as installing if devtron is installing with cicd module (#2576)
- fix: updated wiring for TelemetryEventClientExtended.go (#2588)
- fix: sql script semicolon fix (#2590)
- fix: deploymentApp creation check update (#2600)
- fix: health fix for nodes in resource tree for missing state (#2606)
- fix: Some Helm apps not visible fix (#2538)
- fix: hotfix for bitbucket panic  (#2553)
- fix: schema.json of reference charts update  (#2572)
- fix: ingress new-legacy issue (#2583)
- fix: fixed migration query in 90th verison (#2586)
- fix: Ingress pathtype if condition update (#2592)
- perf: Enforcer RBAC Cache Override Fix (#2608)
## Enhancements
- feat: added support for cd pipelines bulk deletion (#2544)
- feat: Bulk build trigger (#2546)
- feat: App workflow and env overrides Clone and Delete api. (#2565)
- feat: Provide description for Chart Types (#2585)
- feat: Build With Dockerfile template and without dockerfile(using buildpack) (#2562)
- feat: deployment release counter exposed in post stage (#2555)
## Documentation
- docs: Added buildpack doc (#2543)
- docs: Modified the devtron install cmd with gcp cloud (#2580)
- docs: updated doc for aws secret manager  (#2591)
## Others
- chore: Updated issue and pr template (#2570)
- chore: Created update-release-notes.yml github action (#2587)
- feat: Ci Build config API backward compatibilty (#2598)
- chore: Release action update for new release process (#2610)
- chore: Created release.txt for comparison in nightly releases (#2611)
