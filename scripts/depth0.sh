#!/bin/sh

export NAMESPACE=nonsense

for i in {0..24}; do
    kubectl -n $NAMESPACE exec -it swarm-private-$i -- ./geth attach /root/.ethereum/bzzd.ipc --exec="console.log(bzz.hive)" | grep DEPTH
done
