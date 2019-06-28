#!/bin/bash
gcloud builds submit . \
    --config=./chat-clients/cloudbuild.local.yaml >chat-clients-build.log 2>&1 & 
gcloud builds submit . \
    --config=./dice-server/cloudbuild.local.yaml >dice-server-build.log 2>&1 & 
gcloud builds submit . \
    --config=./letsencrypt/cloudbuild.local.yaml >lets-encrypt-build.log 2>&1 &
gcloud builds submit . \
    --config=./www/cloudbuild.local.yaml >www-build.log 2>&1 &
kubectl apply -f ./config/k8s/ >config.log 2>&1 &
wait
kubectl delete pods --all \
&& kubectl get pods --watch