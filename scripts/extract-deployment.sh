#!/bin/sh

NAMESPACE=tony

for i in {0..25}; do
  mkdir -p tony/swarm-private-$i/keystore
  mkdir -p tony/swarm-private-$i/swarm
  kubectl -n $NAMESPACE cp swarm-private-$i:/root/.ethereum/keystore ./tony/swarm-private-$i/keystore &
  kubectl -n $NAMESPACE cp swarm-private-$i:/root/.ethereum/swarm/nodekey ./tony/swarm-private-$i/swarm/nodekey &
done
