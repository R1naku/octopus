#!/bin/bash

for i in {1..500}
do
    curl -s -o /dev/null http://localhost:8080 &
done

wait
echo "end"
