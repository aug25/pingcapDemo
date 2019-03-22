#TiDB cluster test demo

This simple demo use golang to startup tidb cluster from docker, so you need to have docker install first. 
Accept args from cmd line input to specify the number of tidb/pd/tikv nodes, it will automatic startup the cluster with haproxy as well.
Currently, containers are not expose to host, so need to connect to "control" node to run test client

Tested 3 PD, 3 TiKV, 2 TiDB on my MBP 2015

put pingcapDemo into $GOPATH/src
if you want to change node number, please clean /data and /log before run again

## Build test client/server before cluster deploy
I need to do the crossplatform compile, as my host is Mac, containers are Linux

cd crossplatform; ./crossbuild.sh

##Startup tidb cluster
cd Docker; go build DockerServer.go; ./DockerServer -pdno=3 -tikvno=3 -tidbno=2

##Connect to the tidb cluster:
mysql -h 127.0.0.1 -P 3690 -u root 

##Connect to control node:
docker exec -it control sh

##Run test from control node, invoke test on the tidb/pd/tikv node:
cd testbin;
./client 

##Client usage:
\<node> \<operation> : perform test operation on node

exit: exit client

## Supported client operation:
pd1 kill: kill the node pd1

TODO: more test method to be added
Customize test can add to Pd/TiKV/TiDBNodeServer as RPC methond

