#!/bin/bash
   
echo "start follower1"
./main $0
echo "start follower2"
./main $2
echo "start follower3"
./main $3
