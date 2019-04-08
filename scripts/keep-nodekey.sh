#!/bin/sh

NAMESPACE=nonsense

kubectl exec -n $NAMESPACE -ti swarm-private-bootnode-0 -- sh -c "touch /root/.ethereum/delete_datadir && touch /root/.ethereum/keep_nodekey"

for i in {0..39}; do
    kubectl exec -n $NAMESPACE -ti swarm-private-$i -- sh -c "touch /root/.ethereum/delete_datadir && touch /root/.ethereum/keep_nodekey"
done

kubectl -n $NAMESPACE delete pods -l app=swarm-private,component=bootnode,release=swarm-private
kubectl -n $NAMESPACE delete pods -l app=swarm-private,component=swarm,release=swarm-private
