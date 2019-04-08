#!/bin/sh

export NAMESPACE1=nonsense
export NAMESPACE2=tony

for i in {0..39}; do
    echo $i
    kubectl exec -n $NAMESPACE1 -ti swarm-private-$i -- sh -c "cat /root/.ethereum/swarm/nodekey"
    echo " "
    kubectl exec -n $NAMESPACE2 -ti swarm-private-$i -- sh -c "cat /root/.ethereum/swarm/nodekey"
    echo " "
    echo " "
done
