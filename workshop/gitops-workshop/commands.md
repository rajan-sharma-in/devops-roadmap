# GitOps Workshop — One-Page Command Guide

All commands assume you run them from the repo root `gitops-workshop/` on a machine with Docker, kind, kubectl, helm, and gh CLI installed. An `.envrc` (ignored by git) holds local secrets; run `direnv allow` (or `source .envrc`) to load them.

## 0) Load environment (Docker Hub + kubeconfig)
```sh
direnv allow  # or: source .envrc
```

## 1) Create kind cluster with ingress
```sh
kind create cluster --name workshop-cluster --config kind-config-cluster.yaml
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  -n ingress-nginx --create-namespace \
  -f ingress-nginx-values.yaml

kubectl --context kind-workshop-cluster get nodes
kubectl --context kind-workshop-cluster get pods -A | head
```

## 2) Install cert-manager + ARC (in-cluster GitHub runner)
```sh
# cert-manager CRDs and controllers
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.4/cert-manager.crds.yaml
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm upgrade --install cert-manager jetstack/cert-manager \
  -n cert-manager --create-namespace \
  --set installCRDs=false

# Actions Runner Controller (requires PAT in $GITHUB_PAT with repo+workflow+admin:repo_hook scopes)
helm repo add actions-runner-controller https://actions-runner-controller.github.io/actions-runner-controller
helm repo update
helm upgrade --install arc actions-runner-controller/actions-runner-controller \
  -n actions-runner-system --create-namespace \
  --set githubWebhookServer.enabled=false \
  --set authSecret.create=true \
  --set authSecret.github_token=$GITHUB_PAT

# Deploy repo-scoped runner with DinD sidecar
kubectl apply -f runner-rbac.yaml
kubectl apply -f runner-deployment.yaml
kubectl -n actions-runner-system get pods
# Verify runner shows online in GitHub: Settings → Actions → Runners
```

## 3) Set GitHub repo secrets (one time, requires `gh auth login`)

```sh 
cat ~/.kube/config | base64 
```

```sh
gh secret set DOCKERHUB_USERNAME --repo rajan-sharma-in/gitops-workshop --body "$DOCKERHUB_USERNAME"
gh secret set DOCKERHUB_TOKEN    --repo rajan-sharma-in/gitops-workshop --body "$DOCKERHUB_TOKEN"
gh secret set KUBE_CONFIG_B64    --repo rajan-sharma-in/gitops-workshop --body "$KUBE_CONFIG_B64"
```

## 4) Point deployment to your Docker Hub namespace (one time)
Already handled via envsubst in CI; ensure `DOCKERHUB_USERNAME` is set in `.envrc` and GitHub secrets.

## 5) Build/push/deploy via GitHub Actions
```sh
git add .
git commit -m "bootstrap github + dockerhub"  # or your message
git push origin main
```
The workflow `.github/workflows/ci.yml` will:
- Build with Buildx
- Push to Docker Hub (`latest` and `${GITHUB_SHA}`)
- Deploy to the kind cluster using the kubeconfig from `KUBE_CONFIG_B64`

## 6) Validate deployment
```sh
kubectl --context kind-workshop-cluster -n ops-demo get pods
kubectl --context kind-workshop-cluster -n ops-demo rollout status deploy/ops-demo --timeout=120s
kubectl --context kind-workshop-cluster -n ops-demo get svc,ingress,hpa,pdb
```

## 7) Port-forward for app access (if needed)
```sh
kubectl --context kind-workshop-cluster -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
# Then open http://ops.host.docker.internal:8080
```

## 8) Metrics server (needed for HPA)
```sh
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server
helm repo update
helm upgrade --install metrics-server metrics-server/metrics-server \
  -n kube-system \
  --set args="{--kubelet-insecure-tls,--kubelet-preferred-address-types=InternalIP}"
```

## 9) Load testing on/off
```sh
make load-on
make load-off
# verify after load-off that loadgen is 0 and HPA settles back:
kubectl --context kind-workshop-cluster -n ops-demo get deploy ops-demo-loadgen -o jsonpath='{.spec.replicas}{"\n"}'
kubectl --context kind-workshop-cluster -n ops-demo get hpa
```

## 10) Cleanup
```sh
make cleanup          # removes ops-demo namespace resources
kind delete cluster --name workshop-cluster
```
