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

# ---------------------------------------------------------------------------
# Migrations
#
# Two databases, two migration trees, both must be applied:
#   scripts/sql     -> orchestrator DB  (roles, rbac_role_data, rbac_policy_data, ...)
#   scripts/casbin  -> casbin DB        (casbin_rule: the super-admin p-lines)
#
# Connection settings default to this dev setup and are overridden by ../.env
# if present, or by CLI args. Nothing needs passing for the normal case.
# The template cache loads once at boot, so RESTART the orchestrator after
# migrating or new rbac_role_data rows will not be picked up.
# ---------------------------------------------------------------------------
-include ../.env

PG_ADDR         ?= localhost
PG_PORT         ?= 5432
PG_USER         ?= postgres
PG_DATABASE     ?= orchestrator_oss_2
CASBIN_DATABASE ?= casbin_oss_2
STEPS           ?= 1

SQL_DIR    := ../scripts/sql
CASBIN_DIR := ../scripts/casbin
ORCH_DSN    = postgres://$(PG_USER):$(PG_PASSWORD)@$(PG_ADDR):$(PG_PORT)/$(PG_DATABASE)?sslmode=disable
CASBIN_DSN  = postgres://$(PG_USER):$(PG_PASSWORD)@$(PG_ADDR):$(PG_PORT)/$(CASBIN_DATABASE)?sslmode=disable

.PHONY: help apply apply-argo apply-flux namespaces clean clean-argo clean-flux \
        clean-namespaces status check-prereqs argo-crds argo-any-namespace \
        migrate-check migrate-info migrate-up migrate-up-orchestrator migrate-up-casbin \
        migrate-down migrate-down-orchestrator migrate-down-casbin migrate-force

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
	@echo ""
	@echo "Migrations (override PG_* / CASBIN_DATABASE on the CLI):"
	@echo "  migrate-info        show target databases and current versions"
	@echo "  migrate-up          apply both trees (orchestrator + casbin)"
	@echo "  migrate-down        roll back STEPS=1 on both trees"
	@echo "  migrate-force       clear a dirty state: DB=orchestrator|casbin VERSION=n"

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

# ---------------------------------------------------------------------------
# Migration targets
# ---------------------------------------------------------------------------

migrate-check:
	@command -v migrate >/dev/null 2>&1 || { \
	  echo "golang-migrate not found. Install it:"; \
	  echo "  brew install golang-migrate"; exit 1; }

migrate-info: migrate-check
	@echo "orchestrator : $(PG_USER)@$(PG_ADDR):$(PG_PORT)/$(PG_DATABASE)   <- $(SQL_DIR)"
	@echo "casbin       : $(PG_USER)@$(PG_ADDR):$(PG_PORT)/$(CASBIN_DATABASE)   <- $(CASBIN_DIR)"
	@if [ "$(CASBIN_DATABASE)" = "casbin" ]; then \
	  echo ""; \
	  echo "WARNING: CASBIN_DATABASE is the default 'casbin'. On a shared Postgres that"; \
	  echo "         database is shared with every other developer's orchestrator."; \
	  echo "         Set CASBIN_DATABASE=casbin_<yours> in ../.env before migrating."; \
	fi
	@echo ""
	@echo "-- current versions --"
	@printf "orchestrator : "; migrate -path $(SQL_DIR)    -database "$(ORCH_DSN)"   version 2>&1 || true
	@printf "casbin       : "; migrate -path $(CASBIN_DIR) -database "$(CASBIN_DSN)" version 2>&1 || true

migrate-up: migrate-up-orchestrator migrate-up-casbin
	@echo ""
	@echo "Both trees applied. Restart the orchestrator — rbac_role_data and"
	@echo "rbac_policy_data are cached in memory at boot (UserCommonService.go:102)."

migrate-up-orchestrator: migrate-check
	@echo "== orchestrator ($(PG_DATABASE)) =="
	@migrate -path $(SQL_DIR) -database "$(ORCH_DSN)" up

# Forgetting this one is the classic failure: super-admins silently lose access
# to the new resources, because there is no catch-all row in casbin_rule.
migrate-up-casbin: migrate-check
	@echo "== casbin ($(CASBIN_DATABASE)) =="
	@migrate -path $(CASBIN_DIR) -database "$(CASBIN_DSN)" up

# STEPS defaults to 1. Bare `migrate down` rolls back EVERYTHING, so it is never
# used here.
migrate-down: migrate-down-orchestrator migrate-down-casbin

migrate-down-orchestrator: migrate-check
	@echo "== orchestrator ($(PG_DATABASE)) down $(STEPS) =="
	@migrate -path $(SQL_DIR) -database "$(ORCH_DSN)" down $(STEPS)

# Guarded: rolling back the shared 'casbin' database would strip policies from
# every orchestrator pointed at this Postgres. Override with I_KNOW=1.
migrate-down-casbin: migrate-check
	@if [ "$(CASBIN_DATABASE)" = "casbin" ] && [ "$(I_KNOW)" != "1" ]; then \
	  echo "REFUSING: CASBIN_DATABASE=casbin is the shared database."; \
	  echo "  Use your own (CASBIN_DATABASE=casbin_oss_2), or pass I_KNOW=1."; \
	  exit 1; \
	fi
	@echo "== casbin ($(CASBIN_DATABASE)) down $(STEPS) =="
	@migrate -path $(CASBIN_DIR) -database "$(CASBIN_DSN)" down $(STEPS)

# A migration that fails partway leaves the version marked dirty and blocks all
# further migrations. Clear it by forcing to the last known-good version:
#   make migrate-force DB=orchestrator VERSION=36204600
migrate-force: migrate-check
	@test -n "$(VERSION)" || { echo "VERSION=<n> required"; exit 1; }
	@case "$(DB)" in \
	  orchestrator) migrate -path $(SQL_DIR)    -database "$(ORCH_DSN)"   force $(VERSION) ;; \
	  casbin)       migrate -path $(CASBIN_DIR) -database "$(CASBIN_DSN)" force $(VERSION) ;; \
	  *) echo "DB=orchestrator|casbin required"; exit 1 ;; \
	esac
