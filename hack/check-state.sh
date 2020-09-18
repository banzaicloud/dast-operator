#!/bin/sh

state="start"
while true; do
  case "$state" in
  "start")
    echo "initial state"
    state="pending"
    ;;
  "pending")
    echo "pending state"
    kubectl get po -l app.kubernetes.io/component=webhook -n cert-manager -o jsonpath='{.items[*].status.conditions[*].status}' | grep False
    status=$?
    if [ $status != 1 ]; then
      sleep 5
      continue
    fi
    state="running"
    ;;
  "running")
    echo "webhook is running"
    break
    ;;
  *)
    echo "invalid state \"$state\""
    break
    ;;
  esac
done
