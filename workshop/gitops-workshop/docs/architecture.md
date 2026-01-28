# DevOps in Practice: Building & Scaling Systems — Architecture

```mermaid
graph TD
  Laptop[Laptop / Instructor]
  PF[Port-forward\n8080 -> ingress 80\n8443 -> ingress 443]
  Ingress[ingress-nginx]
  GitHub[GitHub\nRepo + Actions]
  Registry[Docker Hub\n<user>/gitops-workshop]
  CI[GitHub Actions\nself-hosted runner]
  Kaniko[Kaniko executor\n(no DinD)]
  App[ops-demo Deployment\n/ /healthz /readyz /metrics]
  SVC[Service ops-demo]
  HPA[HPA CPU 60%\n2-6 replicas]
  PDB[PDB minAvailable=1]
  Loadgen[Load generator\nops-demo-loadgen]

  Laptop --> PF --> Ingress
  Laptop -.git push.-> GitHub
  CI --> Kaniko --> Registry
  CI -->|kubectl apply| App
  CI -->|kubectl apply| HPA
  CI -->|kubectl apply| PDB
  Registry -->|pull image| App
  Ingress --> SVC --> App
  Loadgen -->|HTTP 8080| SVC
  App -->|metrics| HPA
```

- Hostname for app: `ops.host.docker.internal` (ingress on 8080/8443 via port-forward).
- Registry: Docker Hub (`docker.io/<your-username>/gitops-workshop`).
- CI runner uses in-cluster service account + RBAC to deploy to the `ops-demo` namespace.
