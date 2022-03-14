#!/bin/sh
echo "{\"text\":\"What is the melting point of silver?\"}" > input
JSON2=`curl -s -X POST -d @input localhost:3001/alpha`
echo $JSON2

