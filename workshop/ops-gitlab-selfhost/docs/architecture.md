# DevOps in Practice: Building & Scaling Systems — Architecture

```mermaid
graph TD
  Laptop[Laptop / Instructor]
  PF[Port-forward\n8080 -> ingress 80\n8443 -> ingress 443]
  Ingress[ingress-nginx]
  GitLab[GitLab CE\nUI + API]
  Registry[GitLab Registry\ngitlab-registry.gitlab.svc:5000]
  CI[GitLab CI Runner\nK8s executor]
  Kaniko[Kaniko executor\n(no DinD)]
  App[ops-demo Deployment\n/ /healthz /readyz /metrics]
  SVC[Service ops-demo]
  HPA[HPA CPU 60%\n2-6 replicas]
  PDB[PDB minAvailable=1]
  Loadgen[Load generator\nops-demo-loadgen]

  Laptop --> PF --> Ingress --> GitLab
  Laptop -.push/pull.-> Registry
  CI --> Kaniko --> Registry
  CI -->|kubectl apply| App
  CI -->|kubectl apply| HPA
  CI -->|kubectl apply| PDB
  Registry -->|pull image| App
  Ingress --> SVC --> App
  Loadgen -->|HTTP 8080| SVC
  App -->|metrics| HPA
```

- Hostnames: `gitlab.host.docker.internal` (UI), `registry.host.docker.internal` (mapped to registry), `ops.host.docker.internal` (app).
- Traffic enters the cluster via the ingress port-forward on 8080/8443.
- CI runner uses in-cluster service account + RBAC to deploy to the `ops-demo` namespace.
