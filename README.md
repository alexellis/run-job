## run-job

Run a job to completion in Kubernetes.

Creates a Kubernetes Job, waits until it's completed due to success or failure, then gets the Pod logs and deletes the resulting resources.

## Example:

```bash
go run . \ 
    --image ghcr.io/openfaas/config-checker:latest \
    --name checker \
    --namespace openfaas \
    --sa openfaas-checker \
    -out report.txt

```
