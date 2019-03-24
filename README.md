## TiDB cluster test demo

This simple demo use golang to startup tidb cluster from docker, so you need to have docker install first. 
Accept args from cmd line input to specify the number of tidb/pd/tikv nodes, it will automatic startup the cluster with haproxy as well.
Currently, containers are not expose to host, so need to connect to "control" node to run test client

Tested 3 PD, 3 TiKV, 2 TiDB on my MBP 2015

put pingcapDemo into $GOPATH/src

if you want to change node number, please clean /data and /log before run again

## Build test client/server before cluster deploy
I need to do the crossplatform compile, as my host is Mac, containers are Linux

cd crossplatform; ./crossbuild.sh

## Startup tidb cluster
cd Docker; go build DockerServer.go; ./DockerServer -pdno=3 -tikvno=3 -tidbno=2

## Limit container CPU/Memory
\<node> \<resource> \<args> : set resource on node to targe value

tidb0 cpu 2048 50000 250000 : relate weight 2045 compare to other node default 1024, 50000/25000=50% usage

pd0 memory 200000000 500000000: set memory ~200M + swap ~300M on pd0              

## Shutdown and clean cluster
exit: Shutdown and clean cluster

## Connect to the tidb cluster and run SQL:
mysql -h 127.0.0.1 -P 3690 -u root 

## Connect to control node:
docker exec -it control sh

## Run test client from control node, invoke test on the tidb/pd/tikv node:
cd testbin;
./client 

## Client usage:
\<node> \<operation> \<args> : perform test operation on node. can run different operation on different node at the same time, seperate by ';'

tidb0 cpu 10 : stress tidb0 node cpu 100%+ in 10 seconds (work with container cpu restrain can manipulate cpu use percentage, see example)

tikv1 io 10000 : stress tikv1 node disk, write 10GB(10,000 MB)

pd1 kill: kill the node pd1

exit: exit client

Customized test can add to Pd/TiKV/TiDBNodeServer as RPC methond

## Example:

Limit container:

tidb0 memory 600000000 600000000

pd2 cpu 1024 50000 25000

client test:

pd0 kill ; tidb0 cpu 10 ; pd2 cpu 10; tikv0 io 10000

![screenshot](https://github.com/aug25/pingcapDemo/blob/master/screenshot1.png)
![StressIO](https://github.com/aug25/pingcapDemo/blob/master/StressIO.png)


