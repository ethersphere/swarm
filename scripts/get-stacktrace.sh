#!/bin/sh

NAMESPACE=tony

for i in {0..24}; do
    echo $i
    wget -o gor-tony-$i http://localhost:8001/api/v1/namespaces/tony/pods/http:swarm-private-$i:6060/proxy/debug/pprof/goroutine\?debug=2
    echo " "
done
