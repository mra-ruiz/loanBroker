#!/usr/bin/env bash

kind create cluster

kubectl apply -f https://raw.githubusercontent.com/reactive-tech/kubegres/v1.15/kubegres.yaml

# use 'wait' to check for Ready status in .status.conditions[] 
kubectl wait deployment -n kubegres-system kubegres-controller-manager --for condition=Available=True --timeout=90s

kubectl apply -f postgres

# create the table; run my sql script
psql -U username -d myDataBase -a -f myInsertFile
\i path_to_sql_file
