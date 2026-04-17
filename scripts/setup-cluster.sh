#!/usr/bin/env bash

set -e

KUBECTL_VERSION=${1:-""}
# RESOURCE_GROUP_NAME=${2}
# AKS_CLUSTER_NAME=${3}
# ALB_SUBNET_ID=${5}
# ALB_NAMESPACE=${6}
# ALB_NAME=${7}


# Consent Logic
if [[ $# -ge 1 && $1 == "-y" ]]; then
    global_consent=0
else
    global_consent=1
fi


if [[ -z "$RESOURCE_GROUP_NAME" || -z "$AKS_CLUSTER_NAME" || \
      -z "$ALB_SUBNET_ID" || -z "$ALB_NAMESPACE" || -z "$ALB_NAME" ]]; then
    echo "Usage: $0 [-y] <resource_group> <aks_name> <kubectl_version> <alb_subnet_id> <alb_namespace> <alb_name>"
    echo "  -y  Skip confirmation prompts (for automation)"
    exit 1
fi

function assert_consent {
    if [[ $2 -eq 0 ]]; then return 0; fi

    echo -n "$1 [Y/n] "
    read consent
    if [[ ! "${consent}" == "y" && ! "${consent}" == "Y" && ! "${consent}" == "" ]]; then
        echo "Aborted by user."
        exit 1
    fi
}

# Default to no consent for automation, change to 1 for interactive mode
# global_consent=0 

install_dependencies(){
    assert_consent "Add packages necessary to modify your apt-package sources?" ${global_consent}
    export DEBIAN_FRONTEND=noninteractive
    apt-get update
    apt-get install --assume-yes --no-install-recommends apt-transport-https ca-certificates curl gnupg lsb-release wget
    
    mkdir -p -m 755 /etc/apt/keyrings
}

setup_kubectl(){
    if command -v kubectl &>/dev/null; then
        echo "kubectl already installed: $(kubectl version --client --short 2>/dev/null)"
        return 0
    fi

    VERSION=$1
    if [[ -z $VERSION ]]; then
        VERSION=$(curl -L -s https://dl.k8s.io/release/stable.txt | cut -d. -f1-2)
    fi
    
    assert_consent "Add public signing key for the Kubernetes package repositories?" ${global_consent}
    curl -fsSL "https://pkgs.k8s.io/core:/stable:/$VERSION/deb/Release.key" | 
        gpg --dearmor --yes -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
    chmod 644 /etc/apt/keyrings/kubernetes-apt-keyring.gpg

    assert_consent "Add the Kubernetes apt repository to your apt sources?" ${global_consent}
    echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/$VERSION/deb/ /" | 
        tee /etc/apt/sources.list.d/kubernetes.list
}

setup_terraform(){
    assert_consent "Add public signing key for the hashicorp package repositories?" ${global_consent}
    wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor --yes -o /etc/apt/keyrings/hashicorp-archive-keyring.gpg
    chmod 644 /etc/apt/keyrings/hashicorp-archive-keyring.gpg

    assert_consent "Add the Hashicorp apt repository to your apt sources?" ${global_consent}
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" |
        tee /etc/apt/sources.list.d/hashicorp.list
}

setup_azcli() {
    assert_consent "Add Microsoft as a trusted package signer?" ${global_consent}
    curl -sLS https://packages.microsoft.com/keys/microsoft.asc |
        gpg --dearmor --yes -o /etc/apt/keyrings/microsoft.gpg
    chmod 644 /etc/apt/keyrings/microsoft.gpg

    assert_consent "Add the Azure CLI Repository to your apt sources?" ${global_consent}
    
    # Simple logic to find the codename
    CLI_REPO=$(lsb_release -cs)
    
    # Create the list file (This was missing in your original snippet)
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/microsoft.gpg] https://packages.microsoft.com/repos/azure-cli/ $CLI_REPO main" | 
        tee /etc/apt/sources.list.d/azure-cli.list
}

# login_azure_aks(){
#     az login --identity
#     az aks get-credentials --resource-group ${RESOURCE_GROUP_NAME} --name ${AKS_CLUSTER_NAME} --overwrite-existing
#     kubectl config use-context ${AKS_CLUSTER_NAME}
#     kubectl cluster-info
# }

# azure_agfc_addon(){
    
#     # Install Azure CLI extensions.
#     az extension add --name alb
#     az extension add --name aks-preview 
    
#     ALB_ENABLED=$(az aks show --name ${AKS_CLUSTER_NAME} --resource-group ${RESOURCE_GROUP_NAME} --query ["ingressProfile.gatewayApi.installation"] -o tsv)
#     # Update the aks to enable the gateway api and application load balancer
#     if [[ $ALB_ENABLED == "Disabled" || $ALB_ENABLED == "None" ]]; then
#         echo "Updating AKS cluster to enable gateway api and application load balancer"
#         az aks update --name ${AKS_CLUSTER_NAME} --resource-group ${RESOURCE_GROUP_NAME} --enable-gateway-api --enable-application-load-balancer
#     else
#         echo "AKS cluster is already enabled for gateway api and application load balancer"
#     fi

#     echo "Waiting for ALB controller pods to be ready..."
#     kubectl wait pod --selector app=alb-controller -n kube-system --for=condition=Ready 
    
#     # Verify the alb-controller is running in the kube-system namespace
#     echo "Verifying alb-controller is running in the kube-system namespace"
#     kubectl get pods -n kube-system | grep alb-controller || true

#     echo "Waiting for GatewayClass to be accepted..."
#     kubectl wait gatewayclass azure-alb-external --for=condition=Accepted --timeout=600s

#     # Verify GatewayClass azure-alb-external is installed on your cluster
#     echo "Verifying GatewayClass azure-alb-external is installed on your cluster"
#     kubectl get gatewayclass azure-alb-external -o yaml
# }

# azure_agfc_alb(){
#     # 1. Define Namespace
#     KUBECTL_NS=$(cat <<-EOF
# 		apiVersion: v1
# 		kind: Namespace
# 		metadata:
# 		  name: $ALB_NAMESPACE
# 	EOF
#     )

#     # 2. Define ALB Resource
#     KUBECTL_ALB=$(cat <<-EOF
# 		apiVersion: alb.networking.azure.io/v1
# 		kind: ApplicationLoadBalancer
# 		metadata:
# 		  name: $ALB_NAME
# 		  namespace: $ALB_NAMESPACE
# 		spec:
# 		  associations:
# 		  - $ALB_SUBNET_ID
# 	EOF
#     )

#     echo "Creating Namespace: $ALB_NAMESPACE"
#     echo "$KUBECTL_NS" | kubectl apply -f -
    
#     kubectl wait --for=jsonpath='{.status.phase}'=Active namespace/$ALB_NAMESPACE

#     echo "Creating Application Load Balancer: $ALB_NAME"    
#     echo "$KUBECTL_ALB" | kubectl apply -f -

#     echo "Waiting for ALB to be Ready..."
#     kubectl wait --for='jsonpath={.status.conditions[?(@.type=="Deployment")].reason}=Ready' \
#         applicationloadbalancer/$ALB_NAME -n $ALB_NAMESPACE --timeout=600s
# }


setup(){
    setup_kubectl $KUBECTL_VERSION
    setup_terraform
    setup_azcli

    apt-get update
    apt-get install --assume-yes --no-install-recommends kubectl terraform azure-cli

    echo -e "\n--- Verifying Installations ---"
    kubectl version --client
    terraform --version
    az --version
}

# --- Execution ---
# Check for root
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root (use sudo)"
   exit 1
fi

install_dependencies
setup
# login_azure_aks
# azure_agfc_addon
# azure_agfc_alb


# echo ""
# echo "✓ Setup complete"
# echo "  ALB:       $ALB_NAME in namespace $ALB_NAMESPACE"
# echo "  Cluster:   $AKS_CLUSTER_NAME"
# echo "  Run: kubectl get applicationloadbalancer -n $ALB_NAMESPACE"