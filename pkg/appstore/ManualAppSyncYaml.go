package appstore

func manualAppSyncJobByteArr() []byte{
	return []byte(`apiVersion: batch/v1
kind: Job
metadata:
  name: app-manual-sync-job
  namespace: devtroncd
spec:
  template:
    spec:
      containers:
        - name: chart-sync
          image: quay.io/devtron/chart-sync:1227622d-132-3775
          env:
            - name: PG_ADDR
              value: postgresql-postgresql.devtroncd
            - name: PG_DATABASE
              value: orchestrator
            - name: PG_USER
              value: postgres
          envFrom:
            - secretRef:
                name: devtron-secret
      restartPolicy: OnFailure
  backoffLimit: 4
  activeDeadlineSeconds: 15000`)
}
