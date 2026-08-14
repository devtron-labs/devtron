# Test fixtures for Argo CD / Flux CD app-level RBAC.
#
# Creates 4 Argo Applications and 4 Flux Kustomizations across 2 namespaces,
# plus a deliberate Kustomization/HelmRelease name collision.
#
# The multi-cluster dimension comes from the kube context, not the manifests:
# apply the same set to two contexts and register both clusters in Devtron.
#
#   make apply CONTEXT=k3d-dev-1
#   make apply CONTEXT=k3d-dev-2
#   make status CONTEXT=k3d-dev-1
#   make clean  CONTEXT=k3d-dev-1
#
# CONTEXT defaults to the current context. NS1/NS2 are informational — the
# namespaces are hardcoded in the manifests; change them there if needed.

CONTEXT ?= $(shell kubectl config current-context)
KUBECTL  = kubectl --context=$(CONTEXT)
NS1     ?= devtroncd-ent-3
NS2     ?= devtroncd-oss-2

.PHONY: help apply apply-argo apply-flux namespaces clean clean-argo clean-flux \
        clean-namespaces status check-prereqs argo-crds argo-any-namespace

help:
	@echo "Targets (all accept CONTEXT=<kube-context>):"
	@echo "  check-prereqs       verify namespaces and CRDs are present"
	@echo "  namespaces          create the workload namespaces"
	@echo "  argo-crds           install only the Argo Application CRD (list-only testing)"
	@echo "  argo-any-namespace  allow Argo to reconcile apps in NS1/NS2"
	@echo "  apply               apply all Argo + Flux fixtures"
	@echo "  apply-argo          apply the 4 Argo Applications"
	@echo "  apply-flux          apply Flux sources, Kustomizations, collision HelmRelease"
	@echo "  status              list what Devtron will see"
	@echo "  clean               delete everything created here"

check-prereqs:
	@echo "== context: $(CONTEXT)"
	@for ns in $(NS1) $(NS2); do \
	  $(KUBECTL) get ns $$ns >/dev/null 2>&1 \
	    && echo "  ns $$ns: ok" \
	    || echo "  ns $$ns: MISSING - create it or edit the manifests"; \
	done
	@$(KUBECTL) get crd applications.argoproj.io >/dev/null 2>&1 \
	  && echo "  Argo Application CRD: ok" \
	  || echo "  Argo Application CRD: MISSING - run 'make argo-crds' or install Argo CD"
	@$(KUBECTL) get crd kustomizations.kustomize.toolkit.fluxcd.io >/dev/null 2>&1 \
	  && echo "  Flux Kustomization CRD: ok" \
	  || echo "  Flux Kustomization CRD: MISSING - run 'flux install'"
	@$(KUBECTL) get crd helmreleases.helm.toolkit.fluxcd.io >/dev/null 2>&1 \
	  && echo "  Flux HelmRelease CRD: ok" \
	  || echo "  Flux HelmRelease CRD: MISSING - run 'flux install'"

# Devtron reads Application CRs through the k8s API and never contacts argocd-server,
# so the listing works with the CRD alone. Detail/manifest/terminal will NOT work,
# because health, sync status and the managed-resource list are parsed out of
# .status, which only a running controller populates.
argo-crds:
	$(KUBECTL) apply -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/crds/application-crd.yaml

# Without this, Argo's controller ignores Applications outside its own namespace,
# so they list with blank status and an empty resource tree.
argo-any-namespace:
	$(KUBECTL) -n argocd patch cm argocd-cmd-params-cm --type merge \
	  -p '{"data":{"application.namespaces":"$(NS1),$(NS2)"}}'
	$(KUBECTL) -n argocd rollout restart deploy/argocd-server deploy/argocd-application-controller

apply: namespaces apply-argo apply-flux

# Workload namespaces. Argo creates its own via CreateNamespace=true, but Flux
# does not create targetNamespace, so these must exist first.
namespaces:
	$(KUBECTL) apply -f namespaces.yaml

apply-argo: namespaces
	$(KUBECTL) apply -f argo/

# Sources first — a Kustomization whose sourceRef is missing stays in a failed
# state (it still appears in the listing, but with no resource tree).
apply-flux: namespaces
	$(KUBECTL) apply -f flux/gitrepo-ns1.yaml -f flux/gitrepo-ns2.yaml -f flux/helmrepo-ns1.yaml
	$(KUBECTL) apply -f flux/app1ns1.yaml -f flux/app1ns2.yaml -f flux/app2ns1.yaml -f flux/app2ns2.yaml
	$(KUBECTL) apply -f flux/collision-helmrelease-ns1.yaml

status:
	@echo "== Argo Applications (name/namespace = the RBAC object) =="
	@$(KUBECTL) get applications.argoproj.io -A \
	  -o custom-columns=NAMESPACE:.metadata.namespace,NAME:.metadata.name,SYNC:.status.sync.status,HEALTH:.status.health.status 2>/dev/null \
	  || echo "  none / CRD missing"
	@echo
	@echo "== Flux Kustomizations =="
	@$(KUBECTL) get kustomizations -A 2>/dev/null || echo "  none / CRD missing"
	@echo
	@echo "== Flux HelmReleases (collision fixture) =="
	@$(KUBECTL) get helmreleases -A 2>/dev/null || echo "  none / CRD missing"

# Order matters: delete the CRs first so the controllers stop reconciling, then
# drop the workload namespaces. Deleting namespaces first would leave Argo and
# Flux fighting to recreate resources.
clean: clean-argo clean-flux clean-namespaces

# The Applications carry no finalizer, so deleting the CR does not cascade-delete
# what the controller deployed — clean-namespaces handles that.
clean-argo:
	-$(KUBECTL) delete -f argo/ --ignore-not-found

clean-flux:
	-$(KUBECTL) delete -f flux/collision-helmrelease-ns1.yaml --ignore-not-found
	-$(KUBECTL) delete -f flux/app1ns1.yaml -f flux/app1ns2.yaml -f flux/app2ns1.yaml -f flux/app2ns2.yaml --ignore-not-found
	-$(KUBECTL) delete -f flux/gitrepo-ns1.yaml -f flux/gitrepo-ns2.yaml -f flux/helmrepo-ns1.yaml --ignore-not-found

clean-namespaces:
	-$(KUBECTL) delete ns -l rbac-test=true --ignore-not-found
