# Command Cheatsheet (copy/paste friendly)
Source of truth for student-facing commands. Keep this short and accurate.

## Phase 0: GitLab bootstrap (empty instance already running in-cluster)
Goal: create group + project, register runner, wire local repo. GitLab is reached via ingress + port-forward.

1) Get initial root password:
```sh
kubectl --context kind-ops -n gitlab get secret gitlab-gitlab-initial-root-password -ojsonpath='{.data.password}' | base64 -d; echo
```
2) Start ingress port-forward (keep it running):
```sh
kubectl --context kind-ops -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
```
3) Log in to https://gitlab.host.docker.internal:8443 as `root`, change the password.
4) Create a PAT with scopes `api,read_registry,write_registry`:
```sh
export GITLAB_PAT=<your-token>
export GITLAB_API=https://gitlab.host.docker.internal:8443/api/v4
```
5) Create group and project:
```sh
curl --header "PRIVATE-TOKEN: $GITLAB_PAT" \
  --data "name=workshop&path=workshop&visibility=private" \
  "$GITLAB_API/groups"

curl --header "PRIVATE-TOKEN: $GITLAB_PAT" \
  --data "name=ops-workshop-stateless&path=ops-workshop-stateless&namespace_id=$(curl -s --header \"PRIVATE-TOKEN: $GITLAB_PAT\" $GITLAB_API/groups?search=workshop | jq '.[0].id')" \
  "$GITLAB_API/projects"
```
6) (Optional) Add SSH key:
```sh
ssh-keygen -t ed25519 -C "workshop" -f ~/.ssh/gitlab_workshop -N ""
curl --header "PRIVATE-TOKEN: $GITLAB_PAT" \
  --data-urlencode "title=workshop-key" \
  --data-urlencode "key=$(cat ~/.ssh/gitlab_workshop.pub)" \
  "$GITLAB_API/user/keys"
# then: git remote set-url origin git@gitlab.host.docker.internal:workshop/ops-workshop-stateless.git
```
7) Register in-cluster runner (if not already):
```sh
export RUNNER_REG_TOKEN=<from GitLab UI: Admin → Runners>
helm repo add gitlab https://charts.gitlab.io
helm repo update
helm upgrade --install gitlab-runner gitlab/gitlab-runner \
  -n gitlab-runner --create-namespace \
  -f gitlab-runner-values.yaml \
  --set runnerRegistrationToken=$RUNNER_REG_TOKEN
kubectl --context kind-ops -n gitlab-runner get pods
```
8) Wire repo and push:
```sh
git init
git checkout -b main
git remote add origin https://gitlab.host.docker.internal:8443/workshop/ops-workshop-stateless.git
git add .
git commit -m "bootstrap workshop"
git push -u origin main
```

## Phase 1: Cluster + Ingress (kind)
Hosts entry:
```sh
sudo sh -c 'echo "127.0.0.1 gitlab.host.docker.internal registry.host.docker.internal ops.host.docker.internal" >> /etc/hosts'
```
Create multi-node kind cluster (control-plane + 2 workers) and install ingress-nginx:
```sh
kind create cluster --name ops --config kind-config.yaml
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  -n ingress-nginx --create-namespace \
  --set controller.ingressClassResource.default=true \
  --set controller.publishService.enabled=true
kubectl --context kind-ops get nodes
kubectl --context kind-ops get pods -A
kubectl --context kind-ops get svc -n ingress-nginx
```
Cleanup:
```sh
kind delete cluster --name ops
```

## Phase 2: GitLab CE + Runner + Registry (in-cluster)
Install GitLab with provided values:
```sh
helm repo add gitlab https://charts.gitlab.io
helm repo update
helm upgrade --install gitlab gitlab/gitlab -n gitlab --create-namespace -f gitlab-values.yaml
kubectl --context kind-ops -n gitlab get secret gitlab-gitlab-initial-root-password -ojsonpath='{.data.password}' | base64 -d; echo
```
Keep the ingress port-forward from Phase 0 running; access UI at https://gitlab.host.docker.internal:8443.

## Phase 3: App build + deploy (stateless ops-demo)
Local image workflow (optional):
```sh
docker build -t ops-demo:dev .
kind load docker-image ops-demo:dev --name ops
kubectl --context kind-ops -n ops-demo set image deployment/ops-demo ops-demo=ops-demo:dev
```
CI pipeline does Kaniko build + push to internal registry + deploy:
- Push to `main` to trigger pipeline (`build` → `deploy` stages).
- Ingress host: http://ops.host.docker.internal:8080 (via port-forward 8080/8443 to ingress-nginx).

## Quick port-forward command (keep running)
```sh
kubectl --context kind-ops -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
```
