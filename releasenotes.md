<!--upgrade-prerequisites-required-->
> **= = = = = IMPORTANT = = = = =**

IF YOU ARE UPGRADING FROM v0.6.24 OR BELOW VERSIONS, THEN RUN THE FOLLOWING COMMANDS AS A PRE-REQUISITE BEFORE UPGRADE:
```
export RELEASE_NAME=devtron
kubectl label clusterrole kubewatch "app.kubernetes.io/managed-by=Helm" --overwrite
kubectl annotate clusterrole kubewatch "meta.helm.sh/release-name=$RELEASE_NAME" "meta.helm.sh/release-namespace=devtroncd" --overwrite
kubectl label clusterrolebinding kubewatch "app.kubernetes.io/managed-by=Helm" --overwrite
kubectl annotate clusterrolebinding kubewatch "meta.helm.sh/release-name=$RELEASE_NAME" "meta.helm.sh/release-namespace=devtroncd" --overwrite
```

> Ignore the message above if you are not using devtron with cicd mode

> Contact Devtron team on [DISCORD](https://discord.devtron.ai) if you have any concerns.
<!--upgrade-prerequisites-required-->

## v0.6.28

## Bugs
- fix: Not able to see workflow name in case of project with uppercase letters (#4776)
- fix: provider nil fields (#4758)
- fix: git material url sanitised (#4742)
- fix: customTagService missing in workflowdagexec (#4741)
- fix: packages synced (#4743)
- fix: wrong registry creds is used in CD stage (PRE/POST) (#4717)
- fix: force delete option fixed on resource browser and app details page (#4646)
- fix: 28feb packet common bug fix (#4724)
- fix: Added argo app stream API response to resource tree API (#4715)
- fix: wrong image scan execution result (#4691)
- fix: 5xx 3.0 (#4578)
- fix: chart repo delete routing (#4692)
- fix: job workflow req volume mount (#4693)
- fix: User api version for old apis (#4685)
- fix: updated rbac for argo apps listing and detail fetch apis (#4678)
- fix: nil pointer issue in manual sync api (#4681)
- fix: add milli unit in memory (#4671)
- fix: resource terminal update api (#4641)
- fix: chartRef fetch in historical deployment triggers (#4654)
- fix:version upgrade (#4557)
- fix: resource tree bad gateway handle (#4613)
- fix: new api added for getting ciPipelineDetails by only ciPipelineId for CLI (#3708)
## Enhancements
- feat: Cosign plugin (#4543)
- feat: added support for GCS  for image exporter plugin (#4625)
- feat: Github Release Plugin  (#4761)
- feat: User defined git repo for GitOps deployment (#4281)
- feat: change deployment type from gitops to non-gitops and vice-versa for chart store apps (#4666)
- feat:version upgrade for authenticator (#4651)
- feat: active inactive user phase 2 (#4624)
- feat: added support for external argocd apps (#4643)
- feat: support build infra configurations properties configurable from ui (#4583)
- perf: adding  support for git cli operations for existing go-git flows (#4602)
- feat: added proxy in github transport (#4652)
- feat: active inactive user phase 1 (#4589)
- feat: added support for multiple gitops configurations on shared clusters (#4622)
- feat: added support for downloading pod logs (#4539)
## Documentation
- doc: pending revisions from PR-4753 (#4754)
- doc: project-level corrections (#4753)
- doc: Lock Deployment Config (#4732)
- doc: Link corrections and Fixes (#4733)
- doc: Navigation and Scaffolding Fixes (#4716)
- docs: Workflow + Skopeo + Approval + Image Tag at CI + Image Tag + Image Digest at CD (#4507)
- docs: Added the FAQ for the issue:Git-sensor PVC- disk full  (#4714)
- doc: Updated PR Validator (#4697)
- doc: Added Build Infra Doc (#4686)
- doc: added troubleshooting guide for exit status 128 (#4657)
- doc: Added redirection with file type (#4675)
- doc: Fixed Redirections and Broken Links (#4634)
- doc: Revamped Security Feature Doc (#4349)
## Others
- misc: Updated new codeowners for migration scripts (#4781)
- chore: nats common lib version upgrade (#4711)
- chore: Refactoring dag app services  (#4612)
- misc: Changed PR template (#4698)
- chore: AppStore Refactoring v3 (#4621)
- chore: Update CODEOWNERS (#4630)
- chore: Refactoring deployment template GitOps & AppStore (#4616)
- chore: clean appListing api (#4628)


