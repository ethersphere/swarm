#!/bin/sh

NAMESPACE=tony

kubectl exec -n $NAMESPACE -ti swarm-private-bootnode-0 -- touch /root/.ethereum/delete_datadir

for i in {0..24}; do
    kubectl exec -n $NAMESPACE -ti swarm-private-$i -- touch /root/.ethereum/delete_datadir
done

kubectl -n $NAMESPACE delete pods -l app=swarm-private,component=bootnode,release=swarm-private
sleep 10
kubectl -n $NAMESPACE delete pods -l app=swarm-private,component=swarm,release=swarm-private
