## v0.6.15

## Bugs
- fix: updated condition for handling linked ci creation with parent ci docker config overriden (#3243)
- fix: updated condition for getting role filters  (#3261)
- fix: bulk cd deployment trigger failing (#3233)
- fix: Pipeline strategy empty check (#3256)
- fix: SSO Login credentials are sent null to secret if we save after obfuscation (#3248)
- fix: helm pipeline force delete check update (#3245)
- fix: git sensor initContainer indentation fix (#3229)
- fix: updated rbac for devtron apps/helm pipeline (#3227)
- fix: updated query for getting workflow runner by appId and envId for active pipelines only (#3222)
- Fix: App grouping optimisation on ci pipelines (#3219)
- fix: label key can be saved without value if saved without propagation (#3190)
- fix: rbac optimisation for creation of policies for devtron app, helm app and cluster (#2948)
- fix: handled case for failed type assertion (#3208)
## Enhancements
- Feat: Added observability for logged in user (#3204)
## Others
- chore: Discord notification for pager duty label issues  (#3265)
- fix  - handled nil pointer access, fixed not null constraint for git material deletion (#3199)


