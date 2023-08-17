## run-job 🏃‍♂️

[![Sponsor this](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&link=https://github.com/sponsors/alexellis)](https://github.com/sponsors/alexellis)
[![Github All Releases](https://img.shields.io/github/downloads/alexellis/run-job/total.svg)]()
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![build](https://github.com/alexellis/run-job/actions/workflows/build.yml/badge.svg)](https://github.com/alexellis/run-job/actions/workflows/build.yml)

The easiest way to run a simple one-shot job on Kubernetes.

run-job 🏃‍♂️ does the following with a simple YAML file definition:
* Creates a Kubernetes Job
* Watches until it passes or fails
* Collects its logs (if available)
* Deletes the job

It's primary usecase is for [checking OpenFaaS installations for customers](https://github.com/openfaas/config-checker) where it requires a service account to access various resources in a controlled way.

## Examples

The first example is a real-world job for OpenFaaS customers, you probably won't run this example yourself, but read over it to learn the syntax and options. Then feel free to try Example 2 and 3, which anyone should be able to run.

The `image:` field in the Job YAML is for a container image that can be pulled by the cluster.

> Note: the examples in this repo are built with the `faas-cli publish` command because it can create multi-arch container images that work on PCs and ARM devices. You can build your images however you like, or by manually typing in various buildx commands for multi-arch.

### Example 1 - a customer diagnostics tool with a service account

Create a `job.yaml` file:

```yaml
name: checker
image: ghcr.io/openfaas/config-checker:latest
namespace: openfaas
sa: openfaas-checker
# optionally specify resource requests/limits
# requests:
#   cpu: 1000m
# limits:
#   cpu: 2000m
#   nvidia.com/gpu: 1
```

Download run-job from [the releases page](https://github.com/alexellis/run-job/releases), or use arkade:

```bash
$ arkade get run-job
```

Then start the job defined in `job.yaml` and export the logs to a `report.txt` file:

```bash
$ run-job \
    -f job.yaml \
    -out report.txt
```


### Example 2 - kubectl with RBAC

In order to access the K8s API, [an RBAC file](/examples/kubectl/rbac.yaml) is required along with a `serviceAccount` field in the job YAML.

The command `kubectl get nodes -o wide` configured [in the job's YAML file](/examples/kubectl/kubectl_get_nodes_job.yaml).

```bash
$ kubectl apply ./examples/kubectl/rbac.yaml
$ run-job -f ./examples/kubectl/kubectl_get_nodes_job.yaml

Created job get-nodes.default (4097ed06-9422-41c2-86ac-6d4a447d10ab)
....
Job get-nodes.default (4097ed06-9422-41c2-86ac-6d4a447d10ab) succeeded
Deleted job get-nodes

Recorded: 2022-09-05 21:43:57.875629 +0000 UTC

NAME           STATUS   ROLES                       AGE   VERSION        INTERNAL-IP    EXTERNAL-IP   OS-IMAGE                         KERNEL-VERSION   CONTAINER-RUNTIME
k3s-server-1   Ready    control-plane,etcd,master   25h   v1.24.4+k3s1   192.168.2.1   <none>        Raspbian GNU/Linux 10 (buster)   5.10.103-v7l+      containerd://1.6.6-k3s1
k3s-server-2   Ready    control-plane,etcd,master   25h   v1.24.4+k3s1   192.168.2.2   <none>        Raspbian GNU/Linux 10 (buster)   5.10.103-v7l+      containerd://1.6.6-k3s1
k3s-server-3   Ready    control-plane,etcd,master   25h   v1.24.4+k3s1   192.168.2.3   <none>        Raspbian GNU/Linux 10 (buster)   5.10.103-v7l+    containerd://1.6.6-k3s1
```

### Example 3 - light relief with ASCII cows

See also: [examples/cows/Dockerfile](/examples/cows/Dockerfile)

cows.yaml:

```yaml
$ cat <<EOF > cows.yaml
# Multi-arch image for arm64, amd64 and armv7l
image: alexellis2/cows:2022-09-05-1955
name: cows
EOF
```

Run the job:

```bash
$ run-job -f cows.yaml

        ()  ()
         ()()
         (oo)
  /-------UU
 / |     ||
*  ||w---||
   ^^    ^^
Eh, What's up Doc?
```

## Why does this tool exist?

Running a Job in Kubernetes is confusing:

* The spec is very different to what we're used to building (Pods/Deployments)
* The API is harder to use to check if things worked since it uses conditions
* Getting the name of the Pod created by a job is a pain
* Getting the logs from a job is a pain, and needs multiple get/describe/logs commands

Inspired by:

* [alexellis/jaas](https://github.com/alexellis/jaas) built in 2017, now deprecated for running jobs on Docker Swarm
* [stefanprodan/kjob](https://github.com/stefanprodan/kjob) by Stefan Prodan, now unmaintained for 3 years

## Can I get a new option / field / feature?

Raise an issue and explain why you need it and whether it's for work or pleasure.

PRs will not be approved prior to an issue being created and agreed upon.

License: MIT
