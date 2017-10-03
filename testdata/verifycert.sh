#!/bin/bash
kubectl get certificates -o json | jq .items[0].status.certificate | sed 's/[\"]//g' | base64 -D |  openssl x509  -text -noout
