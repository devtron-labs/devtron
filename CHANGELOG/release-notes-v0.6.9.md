## v0.6.9

## Bugs
- fix: specific error message if token is invalid while saving cluster (#2760)
## Enhancements
- feat: Resource status for integration (#2620)
- feat: Cluster terminal access for Super Admins (#2654)
- feat: Dynamic handling of deployment templates (#2596)
## Documentation
- docs: description for charts (#2646)


## v0.6.9-rc.1

## Bugs
- fix: Ignoring propagating invalid labels in app to values.yaml during deployment (#2734)
- fix: app clone response update for empty material and templates (#2717)
- fix: getting error message on deleting linked ci pipeline (#2732)
## Enhancements
- feat: Gitops or helm options for cd (#2673)
- feat: added BE support for allowing insecure tls connection in gitOps (#2738)
## Documentation
- docs: minor updates (#2)
- docs: webhook updates links (#2716)
- docs: webhook ci documentation updates (#2637)
- docs: add getting started doc (#2658)
- docs: broken links fixed (#2741)
## Others
- minor updates (#2722)
- fix for app type setting (#2723)
- task: updated environment variables in pre/post cd workflow request (#2727)
- Revert "feat: added BE support for allowing insecure tls connection in gitOps (#2738)" (#2739)
- minor updates (#2742)


## v0.6.9-rc.0

## Bugs
- fix: removed required validation for deployment template (#2713)
- fix: env variable tag bug in ci cd config (#2698)
- fix: Block deletion of devtronapp(devtron-operator chart)  in helm apps (#2701)
- fix: migration down sql for 84 (#2706)
- Fix: Skip external ci webhook clone while app clone (#2704)
- fix: deployment template editor view support in create app api (#2697)
- Fix:  External ci webhook response for api token. (#2699)
- fix: wf deletion bug and bulk cd pipeline req (#2693)
- fix: updated cluster connection status for update api (#2686)
## Enhancements
- Feat: External Ci webhook new interface, Deployment separated from build node. (#2664)
## Documentation
- docs: external-links-docs (#2679)
## Others
- sql migration error fix (#2695)
- External links extention for app level  (#2573)
- added support for giving names in wf bulk deletion req (#2690)


