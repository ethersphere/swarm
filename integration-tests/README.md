# Integration tests for various features of Swarm

This folder contains the test environments, as well as the integration tests parameters for various features of Swarm

## Pushsync

1. Make sure you amend the Docker image version you want to use for a given run of the push-sync integration tests, at:

* `pushsync/deployment/version.yaml`

* `pushsync/smoke-job.yaml`

2. Setup a private deployment for a push-sync integration test

```
cd pushsync
helmsman -apply -f deployment.yaml
```

3. Apply a smoke test job:

```
cd pushsync
kubectl apply -f smoke-job.yaml -n pushsync
```


### Teardown

1. Remove the `swarm` deployment

```
helm del --purge swarm-private --tiller-namespace pushsync
```

2. Remove the smoke test job

```
kubectl -n pushsync delete job smoke --force --grace-period=0
```

3. Remove all PVCs and volumes

```
kubectl -n pushsync get pvc | awk '{print $1}' | xargs kubectl -n pushsync delete pvc
```
