#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BOOTSTRAP_FILES_DIR="$SCRIPT_DIR/bootstrap-files"
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
        if [[ "$(uname)" == "Darwin" ]]; then
            open -a Docker
        else
            systemctl --user start docker-desktop 2>/dev/null || systemctl start docker 2>/dev/null || echo "Warning: Could not auto-start Docker. Please start it manually."
        fi
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
    
    kind create cluster --config "$BOOTSTRAP_FILES_DIR/kind-config.yaml" --name "$CLUSTER_NAME"
    
    echo "Cluster created successfully!"
    kubectl get nodes
    echo ""
}

# Install flux-operator
install_flux_operator() {
    echo "=== Installing Flux Operator ==="
    
    kubectl apply -f https://github.com/controlplaneio-fluxcd/flux-operator/releases/latest/download/install.yaml
    
    echo "Waiting for Flux Operator to be ready..."
    kubectl wait --for=condition=available --timeout=120s deployment/flux-operator -n flux-system
    
    echo "Flux Operator installed successfully!"
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

# Create FluxInstance
create_flux_instance() {
    echo "=== Creating FluxInstance ==="
    
    kubectl apply -f "$BOOTSTRAP_FILES_DIR/flux-instance.yaml"
    
    echo "Waiting for FluxInstance to be ready..."
    kubectl wait --for=condition=ready --timeout=180s fluxinstance/flux -n flux-system
    
    echo "FluxInstance created successfully!"
    kubectl get fluxinstances -n flux-system
    echo ""
    
    echo "Waiting for Flux controllers to be ready..."
    kubectl wait --for=condition=available --timeout=120s deployment -n flux-system -l app.kubernetes.io/part-of=flux
    
    echo "Flux controllers are ready!"
    kubectl get deployments -n flux-system
    echo ""
}

# Wait for LGTM stack deployment
wait_for_lgtm() {
    echo "=== Waiting for LGTM Stack Deployment ==="
    echo "This may take several minutes..."
    echo ""
    
    echo "Waiting for HelmRepositories..."
    kubectl wait --for=condition=ready --timeout=120s helmrepository/grafana -n flux-system
    kubectl wait --for=condition=ready --timeout=120s helmrepository/prometheus-community -n flux-system
    echo "HelmRepositories are ready!"
    echo ""
    
    echo "Waiting for HelmReleases..."
    for release in mimir loki tempo grafana alloy kube-prometheus-stack; do
        namespace="observability"
        if [ "$release" = "kube-prometheus-stack" ]; then
            namespace="monitoring"
        fi
        echo "  - Waiting for $release..."
        kubectl wait --for=condition=ready --timeout=300s helmrelease/$release -n $namespace 2>/dev/null || echo "    $release may still be installing..."
    done
    echo ""
}

# Verify setup
verify_setup() {
    echo "=== Verification ==="
    echo ""
    echo "Cluster Nodes:"
    kubectl get nodes
    echo ""
    echo "Flux Operator:"
    kubectl get deployments -n flux-system -l app.kubernetes.io/name=flux-operator
    echo ""
    echo "Flux Controllers:"
    kubectl get deployments -n flux-system -l app.kubernetes.io/part-of=flux
    echo ""
    echo "FluxInstance:"
    kubectl get fluxinstance -n flux-system
    echo ""
    echo "GitRepository:"
    kubectl get gitrepositories -n flux-system
    echo ""
    echo "Namespaces:"
    kubectl get namespaces | grep -E "observability|monitoring"
    echo ""
    echo "HelmReleases:"
    kubectl get helmreleases -A 2>/dev/null || echo "  No HelmReleases found yet - still deploying..."
    echo ""
}

# Main execution
main() {
    check_prerequisites
    create_cluster
    install_flux_operator
    configure_ssh
    create_flux_instance
    wait_for_lgtm
    verify_setup
    
    echo "=== Bootstrap Complete! ==="
    echo ""
    echo "Your cluster is ready with the full LGTM stack!"
    echo ""
    echo "Access Grafana:"
    echo "  kubectl port-forward svc/grafana 3000:3000 -n observability"
    echo "  Then open http://localhost:3000"
    echo ""
    echo "To use the cluster:"
    echo "  kubectl cluster-info --context kind-$CLUSTER_NAME"
}

main "$@"
