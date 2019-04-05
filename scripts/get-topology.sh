#!/bin/sh

NAMESPACE=tony

for i in {0..24}; do
    echo $i
    kubectl exec -n $NAMESPACE -ti swarm-private-$i -- ./geth attach /root/.ethereum/bzzd.ipc --exec="console.log(bzz.hive)"
    echo " "
done
