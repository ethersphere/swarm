#!/bin/sh

for i in {0..100}; do
  echo $i && echo " " && kubectl exec -n tony -ti swarm-private-$i -- du -sh /root/.ethereum/swarm
done
wait
