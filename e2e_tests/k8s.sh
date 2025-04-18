#!/bin/bash
# shellcheck disable=SC2046

echo Building images
docker build -t smocker ./smocker
docker build -t venom ./venom
docker build -t villip ..
docker build -t tcpserver ./tcp/server
docker build -t tcpclient ./tcp/client

echo Creating cluster
kind create cluster --name villip --wait 30s

kind --name villip load docker-image smocker venom villip tcpserver tcpclient

echo Creating workload
kubectl --context kind-villip create cm venom-tests --from-file=tests
kubectl --context kind-villip create cm smocker-mocks --from-file=mocks
kubectl --context kind-villip create cm villip-conf --from-file=villip

kubectl --context kind-villip apply -f manifests/tcp.yaml
kubectl --context kind-villip apply -f manifests/smocker.yaml
kubectl --context kind-villip apply -f manifests/villip.yaml

echo Let time to smocker container to load the mocks and tcp test to finish
sleep 30

echo Running tests
kubectl --context kind-villip apply -f manifests/venom.yaml
