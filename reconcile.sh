#!/usr/bin/env bash
set -euo pipefail

echo "Reconciling FluxInstance..."
flux-operator -n flux-system reconcile resource FluxInstance/flux

echo "Reconciling Kustomizations..."
for ks in $(kubectl -n flux-system get kustomization -o jsonpath='{.items[*].metadata.name}'); do
  echo "  Kustomization/$ks"
  flux-operator -n flux-system reconcile resource "Kustomization/$ks"
done

echo "Reconciling HelmReleases..."
for hr in $(kubectl get helmrelease --all-namespaces -o json \
  | jq -r '.items[] | "\(.metadata.namespace) \(.metadata.name)"'); do
  ns="${hr%% *}"
  name="${hr##* }"
  echo "  HelmRelease/$name ($ns)"
  flux-operator -n "$ns" reconcile resource "HelmRelease/$name"
done

echo "Reconciliation triggered for all resources."
