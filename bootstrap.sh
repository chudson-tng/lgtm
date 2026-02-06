#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLUSTER_NAME="${CLUSTER_NAME:-lgtm-cluster}"
REPO_URL="${REPO_URL:-git@github.com:chudson-tng/lgtm.git}"
SSH_KEY_PATH="${SSH_KEY_PATH:-$HOME/.ssh/id_ed25519}"
SSH_PUB_KEY_PATH="${SSH_KEY_PATH}.pub"

echo "=== LGTM Cluster Bootstrap Script ==="
echo "Cluster name: $CLUSTER_NAME"
echo "Repository: $REPO_URL"
echo "SSH Key: $SSH_KEY_PATH"
echo ""

# Check prerequisites
check_prerequisites() {
    echo "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        echo "Error: Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        echo "Starting Docker Desktop..."
        open -a Docker
        echo "Waiting for Docker to be ready..."
        for i in {1..30}; do
            if docker info &> /dev/null; then
                echo "Docker is ready!"
                break
            fi
            sleep 2
        done
        if ! docker info &> /dev/null; then
            echo "Error: Docker failed to start"
            exit 1
        fi
    fi
    
    if ! command -v kind &> /dev/null; then
        echo "Error: kind is not installed. Install with: brew install kind"
        exit 1
    fi
    
    if ! command -v kubectl &> /dev/null; then
        echo "Error: kubectl is not installed. Install with: brew install kubectl"
        exit 1
    fi
    
    if ! command -v flux &> /dev/null; then
        echo "Installing flux CLI..."
        curl -s https://fluxcd.io/install.sh | bash
    fi
    
    if [ ! -f "$SSH_KEY_PATH" ]; then
        echo "Error: SSH private key not found at $SSH_KEY_PATH"
        exit 1
    fi
    
    if [ ! -f "$SSH_PUB_KEY_PATH" ]; then
        echo "Error: SSH public key not found at $SSH_PUB_KEY_PATH"
        exit 1
    fi
    
    echo "All prerequisites met!"
    echo ""
}

# Create Kind cluster
create_cluster() {
    echo "=== Creating Kind Cluster ==="
    
    if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
        echo "Cluster '$CLUSTER_NAME' already exists. Deleting it first..."
        kind delete cluster --name "$CLUSTER_NAME"
    fi
    
    cat > "$SCRIPT_DIR/kind-config.yaml" <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
  - role: worker
  - role: worker
EOF
    
    kind create cluster --config "$SCRIPT_DIR/kind-config.yaml" --name "$CLUSTER_NAME"
    
    echo "Cluster created successfully!"
    kubectl get nodes
    echo ""
}

# Install Flux CD
install_flux() {
    echo "=== Installing Flux CD ==="
    
    flux install
    
    echo "Waiting for Flux controllers to be ready..."
    kubectl wait --for=condition=available --timeout=120s deployment -n flux-system -l app.kubernetes.io/part-of=flux
    
    echo "Flux CD installed successfully!"
    kubectl get deployments -n flux-system
    echo ""
}

# Configure SSH access
configure_ssh() {
    echo "=== Configuring SSH Access ==="
    
    # Create known hosts file
    TEMP_DIR=$(mktemp -d)
    ssh-keyscan -t ed25519 github.com > "$TEMP_DIR/github_known_hosts" 2>/dev/null
    
    # Create the SSH secret
    kubectl create secret generic lgtm-repo-ssh \
        --namespace=flux-system \
        --from-file=identity="$SSH_KEY_PATH" \
        --from-file=identity.pub="$SSH_PUB_KEY_PATH" \
        --from-file=known_hosts="$TEMP_DIR/github_known_hosts" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    rm -rf "$TEMP_DIR"
    
    echo "SSH secret created successfully!"
    echo ""
}

# Create GitRepository
create_gitrepository() {
    echo "=== Creating GitRepository ==="
    
    # Convert git@github.com:chudson-tng/lgtm.git to ssh://git@github.com/chudson-tng/lgtm.git
    SSH_URL=$(echo "$REPO_URL" | sed 's|git@github.com:|ssh://git@github.com/|')
    
    cat > "$SCRIPT_DIR/gitrepository.yaml" <<EOF
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: lgtm-repo
  namespace: flux-system
spec:
  interval: 1m0s
  ref:
    branch: main
  url: $SSH_URL
  secretRef:
    name: lgtm-repo-ssh
EOF
    
    kubectl apply -f "$SCRIPT_DIR/gitrepository.yaml"
    
    echo "Waiting for GitRepository to be ready..."
    kubectl wait --for=condition=ready --timeout=60s gitrepository/lgtm-repo -n flux-system
    
    echo "GitRepository created successfully!"
    kubectl get gitrepositories -n flux-system
    echo ""
}

# Verify setup
verify_setup() {
    echo "=== Verification ==="
    echo ""
    echo "Cluster Nodes:"
    kubectl get nodes
    echo ""
    echo "Flux Controllers:"
    kubectl get deployments -n flux-system
    echo ""
    echo "GitRepository:"
    kubectl get gitrepositories -n flux-system
    echo ""
    echo "Recent Events:"
    kubectl get events -n flux-system --field-selector involvedObject.name=lgtm-repo --sort-by='.lastTimestamp' | tail -3
    echo ""
}

# Main execution
main() {
    check_prerequisites
    create_cluster
    install_flux
    configure_ssh
    create_gitrepository
    verify_setup
    
    echo "=== Bootstrap Complete! ==="
    echo ""
    echo "Your cluster is ready with Flux CD configured to pull from:"
    echo "  $REPO_URL"
    echo ""
    echo "Next steps:"
    echo "  1. Add your kustomizations to the repository"
    echo "  2. Create Kustomization resources to deploy them"
    echo ""
    echo "To use the cluster:"
    echo "  kubectl cluster-info --context kind-$CLUSTER_NAME"
}

main "$@"
