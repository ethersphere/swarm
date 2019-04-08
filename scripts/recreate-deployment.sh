#!/bin/sh

export NAMESPACE=tony

echo "Removing all keystore folders"
for i in {0..24}; do
    kubectl exec -n $NAMESPACE -ti swarm-private-$i -- sh -c "rm -rf /root/.ethereum/keystore" &
done

sleep 5

echo "Copying over keystore and nodekey files"
for i in {0..24}; do
    kubectl -n $NAMESPACE cp ./tony/swarm-private-$i/keystore swarm-private-$i:/root/.ethereum &
    kubectl -n $NAMESPACE cp ./tony/swarm-private-$i/swarm/nodekey swarm-private-$i:/root/.ethereum/swarm/nodekey &
done

sleep 10

echo "Adding hooks"
kubectl exec -n $NAMESPACE -ti swarm-private-bootnode-0 -- sh -c "touch /root/.ethereum/delete_datadir && touch /root/.ethereum/keep_nodekey" &

for i in {0..24}; do
    kubectl exec -n $NAMESPACE -ti swarm-private-$i -- sh -c "touch /root/.ethereum/delete_datadir && touch /root/.ethereum/keep_nodekey" &
done

sleep 10

kubectl -n $NAMESPACE delete pods -l app=swarm-private,component=bootnode,release=swarm-private

sleep 10

kubectl -n $NAMESPACE delete pods -l app=swarm-private,component=swarm,release=swarm-private
