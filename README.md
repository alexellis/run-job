## run-job

Run a job to completion in Kubernetes.

Creates a Kubernetes Job, waits until it's completed due to success or failure, then gets the Pod logs and deletes the resulting resources.

## Example:

Create a `job.yaml` file:

```yaml
name: checker
image: ghcr.io/openfaas/config-checker:latest
namespace: openfaas
sa: openfaas-checker
```

```bash
run-job \
    -f job.yaml \
    -out report.txt
```
