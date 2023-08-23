# k8s-fix-unhealthy-pod

This is a simple program to fix the unhealthy pod with error message 'context deadline exceeded (Client.Timeout exceeded while awaiting headers)'

## Usage

create below CronJob to run the program every 5 minutes

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: fix-unhealthy-pod-job
  namespace: kube-system
spec:
  schedule: "*/5 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: fix-unhealthy-pod
              image: ghcr.io/heavenwing/k8s-fix-unhealthy-pod:main
          serviceAccountName: pod-garbage-collector
          restartPolicy: "Never"
```

and deploy it into your Kubernetes cluster.

NOTE: it will check default namespace only, if you want to check other namespaces, you can pass -ns=OTHER_NAMESPACE flag to the program.