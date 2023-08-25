# k8s-fix-unhealthy-pod

This is a simple program to fix the unhealthy pod with error message 'context deadline exceeded (Client.Timeout exceeded while awaiting headers)'

## Usage

create below CronJob to run the program every 5 minutes

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: podfixer-account
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: podfixer-role
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "delete"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: podfixer-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: podfixer-role
subjects:
  - kind: ServiceAccount
    name: podfixer-account
    namespace: default
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: fix-unhealthy-pod-job
  namespace: default
spec:
  schedule: "*/5 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: fix-unhealthy-pod
              image: ghcr.io/heavenwing/k8s-fix-unhealthy-pod:main
          serviceAccountName: podfixer-account
          restartPolicy: "Never"
```

and deploy it into your Kubernetes cluster.

NOTE: it will check default namespace only, if you want to check other namespaces, you can pass -ns=OTHER_NAMESPACE flag to the program.