## run-job 🏃‍♂️ 

The easiest way to run a simple one-shot job on Kubernetes.

run-job 🏃‍♂️ does the following with a simple YAML file definition:
* Creates a Kubernetes Job
* Watches until it passes or fails
* Collects its logs (if available)
* Deletes the job

It's primary usecase is for [checking OpenFaaS installations for customers](https://github.com/openfaas/config-checker) where it requires a service account to access various resources in a controlled way.

## Example

Create a `job.yaml` file:

```yaml
name: checker
image: ghcr.io/openfaas/config-checker:latest
namespace: openfaas
sa: openfaas-checker
```

Run the job and export the logs to a report.txt file:

```bash
run-job \
    -f job.yaml \
    -out report.txt
```

## Why does this tool exist?

Running a Job in Kubernetes is confusing:

* The spec is very different to what we're used to building (Pods/Deployments)
* The API is harder to use to check if things worked since it uses conditions
* Getting the name of the Pod created by a job is a pain
* Getting the logs from a job is a pain, and needs multiple get/describe/logs commands

Inspired by:

* [alexellis/jaas](https://github.com/alexellis/jaas) built in 2017, now deprecated for running jobs on Docker Swarm 
* Also [stefanprodan/kjob](https://github.com/stefanprodan/kjob) by Stefan Prodan, now unmaintained for 3 years

## Can I get a new option / field / feature?

Raise an issue and explain why you need it and whether it's for work or pleasure.

PRs will not be approved prior to an issue being created and agreed upon.

License: MIT
