#!/bin/bash
#
# This script aims to manage k8s clusters

set -o nounset
set -o errexit
#set -o xtrace


function usage() {
    cat <<EOF
Usage: easzctl COMMAND [args]
Cluster-wide operation:
    checkout		To switch to context <clustername>, or create it if not existed
    destroy		To destroy the current cluster, '--purge' to also delete the context
    list		To list all of clusters managed
    setup		To setup a cluster using the current context
    start-aio		To quickly setup an all-in-one cluster for testing (like minikube)
In-cluster operation:
    add-etcd		To add a etcd-node to the etcd cluster
    add-master		To add a kube-master(master node) to the k8s cluster
    add-node		To add a kube-node(work node) to the k8s cluster
    del-etcd		To delete a etcd-node from the etcd cluster
    del-master		To delete a kube-master from the k8s cluster
    del-node		To delete a kube-node from the k8s cluster
    upgrade		To upgrade the k8s cluster
Extra operation:
    basic-auth   	To enable/disable basic-auth for apiserver
Use "easzctl help <command>" for more information about a given command.
EOF
}

function process_cmd() {
    echo -e "[INFO] \033[33m$ACTION\033[0m : $CMD"
    $CMD || { echo -e "[ERROR] \033[31mAction failed\033[0m : $CMD"; return 1; }
    echo -e "[INFO] \033[32mAction successed\033[0m : $CMD"
}

function add-etcd() {
  # check new node's address regexp
  [[ $1 =~ ^(2(5[0-5]{1}|[0-4][0-9]{1})|[0-1]?[0-9]{1,2})(\.(2(5[0-5]{1}|[0-4][0-9]{1})|[0-1]?[0-9]{1,2})){3}$ ]] || { echo "[ERROR] Invalid ip address!"; return 2; }

  # check if the new node already exsited
  sed -n '/^\[etcd/,/^\[kube-master/p' $BASEPATH/hosts|grep "^$1[^0-9]*$" && { echo "[ERROR] etcd $1 already existed!"; return 2; }

  # input an unique NODE_NAME of the node in etcd cluster
  echo "Please input an UNIQUE name(string) for the new node: "
  read -t15 NAME
  sed -n '/^\[etcd/,/^\[kube-master/p' $BASEPATH/hosts|grep "$NAME" && { echo "[ERROR] name [$NAME] already existed!"; return 2; }

  # add a node into 'etcd' group
  sed -i "/\[etcd/a $1 NODE_NAME=$NAME" $BASEPATH/hosts

  # check if playbook runs successfully
  ansible-playbook $BASEPATH/tools/01.addetcd.yml -e NODE_TO_ADD=$1 || { sed -i "/$1 NODE_NAME=$NAME/d" $BASEPATH/hosts; return 2; }

  # restart apiservers to use the new etcd cluster
  ansible-playbook $BASEPATH/04-kube-master.yml -t restart_master || { echo "[ERROR] Unexpected failures in master nodes!"; return 2; }

  # save current cluster context if needed
  # [ -f "$BASEPATH/.cluster/current_cluster" ] && save_context
  return 0
}

### Main Lines ###############################################

BASEPATH=/etc/ansible

[ "$#" -gt 0 ] || { usage >&2; exit 2; }

case $1 in
  (add-etcd)
    [ "$#" -gt 1 ] || { usage >&2; exit 2; }
    ACTION="Action: add a etcd node"
    CMD="add-etcd $2"
  ;;

  (*)
    usage
    exit 0
  ;;
esac

process_cmd
