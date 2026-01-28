# Homework: Deploy Mautic on any Kubernetes cluster

This lab is self-contained: it creates the namespace, database, app, service, and ingress. You just need a working Kubernetes cluster with an Ingress controller.

- Default context: `CTX=${CTX:-your-cluster-context}` (export CTX to override)
- Default host: `mautic.example.com` (change to your DNS/hosts entry)
- Ingress class: `nginx` (change `ingressClassName` in the manifest if you use another controller)

## 1) Prep
```sh
export CTX=<your-kube-context>
kubectl config use-context $CTX
kubectl --context $CTX get nodes
```

## 2) Create namespace + MariaDB (Bitnami)
```sh
kubectl --context $CTX create namespace mautic
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
export MAUTIC_DB_PASS=<set-strong-password>
helm upgrade --install mautic-mariadb bitnami/mariadb -n mautic --kube-context $CTX \
  --set auth.rootPassword=$MAUTIC_DB_PASS \
  --set auth.password=$MAUTIC_DB_PASS \
  --set auth.username=mautic \
  --set auth.database=mautic \
  --set primary.persistence.enabled=true \
  --set primary.persistence.size=8Gi
```

## 3) Deploy Mautic (manifests provided)
Update the host in `deploy-mautic/mautic-app.yaml` if needed (search for `mautic.example.com`). Then apply:
```sh
kubectl --context $CTX apply -f deploy-mautic/mautic-app.yaml
kubectl --context $CTX -n mautic get pods
kubectl --context $CTX -n mautic get svc,ingress
```

## 4) Access
- Create or edit a DNS/hosts entry pointing your host (default `mautic.example.com`) to the ingress controller address.
- If you don't have a LoadBalancer, port-forward the ingress controller:
  ```sh
  kubectl --context $CTX -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
  # Browser: http://mautic.example.com:8080
  ```

## 5) Homework checks
- [ ] All pods in `mautic` namespace are Running.
- [ ] Mautic UI loads via ingress on your chosen host.
- [ ] MariaDB PVCs are Bound and data persists across pod restarts.

Stretch ideas: add TLS via cert-manager; add basic auth or IP allowlist on the ingress; enable cron jobs via `MAUTIC_RUN_CRON_JOBS=true`.
