#!/bin/bash
clear
echo "Username set to 'Student':"

echo "Type in the last byte of the IP to the elvator you want to connect to:"
read IP
echo "Connecting to 129.241.187."$IP
scp -rq /home/student/go/bin/elevator student@129.241.187.$IP:~/top_kek/elevator
ssh student@129.241.187.$IP "top_kek/elevator"

echo Elevator script stopping...
ssh student@129.241.187.$IP "killall elevator"


