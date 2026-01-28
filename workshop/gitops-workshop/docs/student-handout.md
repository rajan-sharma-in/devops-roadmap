# DevOps in Practice: Building & Scaling Systems — Student Handout (commands only)

> **GitHub + Docker Hub quick path (use for this repo)**
> 1) Set GitHub secrets: `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`, `KUBE_CONFIG_B64`.
> 2) Commit + push to `main`; GitHub Actions envsubsts `k8s/02-deployment.yaml`, builds, pushes, and deploys automatically.
> 4) Hit the app at `http://ops.host.docker.internal:8080`.
>
> The steps below are the legacy GitLab bootstrap flow; keep only if you need to run this workshop on a self-hosted GitLab.

## Bootstrap GitLab (fresh instance)
1. Grab initial root password (cluster secret):
   ```bash
   kubectl --context kind-mautic -n gitlab get secret gitlab-gitlab-initial-root-password -ojsonpath='{.data.password}' | base64 -d; echo
   ```
2. Start ingress port-forward (keep running):
   ```bash
   kubectl --context kind-mautic -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
   ```
3. Login to https://gitlab.host.docker.internal:8443 as `root`, change the password.
4. Create a Personal Access Token with scopes `api, read_registry, write_registry` and export it:
   ```bash
   export GITLAB_PAT=<your-token>
   export GITLAB_API=https://gitlab.host.docker.internal:8443/api/v4
   ```
5. Create group + project (names must match the workshop scripts):
   ```bash
   curl --header "PRIVATE-TOKEN: $GITLAB_PAT" --data "name=workshop&path=workshop&visibility=private" "$GITLAB_API/groups"
   GROUP_ID=$(curl -s --header "PRIVATE-TOKEN: $GITLAB_PAT" "$GITLAB_API/groups?search=workshop" | jq '.[0].id')
   curl --header "PRIVATE-TOKEN: $GITLAB_PAT" --data "name=ops-workshop-stateless&path=ops-workshop-stateless&namespace_id=$GROUP_ID" "$GITLAB_API/projects"
   ```
6. Wire your local repo (inside the workshop folder):
   ```bash
   git init
   git checkout -b main
   git remote add origin https://gitlab.host.docker.internal:8443/workshop/ops-workshop-stateless.git
   git add .
   git commit -m "bootstrap workshop"
   git push -u origin main
   ```

## Setup
7. Set hosts (once on laptop):
   ```bash
   echo "127.0.0.1 gitlab.host.docker.internal registry.host.docker.internal ops.host.docker.internal" | sudo tee -a /etc/hosts
   kubectl --context kind-mautic -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
   ```

8. Clone template into your GitLab group (if not already cloned):
   ```bash
   git clone https://gitlab.host.docker.internal:8443/<group>/ops-workshop-stateless.git
   cd ops-workshop-stateless
   ```

9. Run preflight:
   ```bash
   make preflight
   ```

## Build & Deploy
10. Push to main to trigger CI:
   ```bash
   git add . && git commit -m "kickoff" && git push origin main
   ```

11. Watch pipeline and rollout:
   ```bash
   make watch
   ```

## Load generator
12. Turn load on/off:
   ```bash
   make load-on
   make load-off
   ```

## Cleanup
13. Remove workshop resources when done:
   ```bash
   make cleanup
   ```

---

## Checkpoints (expected output snippets)
- App reachable:
  ```bash
  curl -s http://ops.host.docker.internal:8080/
  # {"message":"Hello DevOps Workshop","version":"v1.0.0","pod":"ops-demo-xxxxx"}
  ```

- Pipeline green:
  GitLab CI shows `build` and `deploy` stages passed (green ticks).

- Rollout succeeded:
  ```bash
  kubectl --context kind-mautic -n ops-demo rollout status deploy/ops-demo
  # deployment "ops-demo" successfully rolled out
  ```

- HPA scaled from 2 -> N (while load-on):
  ```bash
  kubectl --context kind-mautic -n ops-demo get hpa
  # NAME       REFERENCE           TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
  # ops-demo   Deployment/ops-demo 75%/60%   2         6         5          3m
  ```
