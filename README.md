

# APIK3S

This is an api developed for communicate with a k3s cluster and automate the deploy of nfs server exposed by the cluster.


## Requirement
This api have some requirement:

 - k3s cluster installed without ServiceLB
 - MetalLB installed
 - Longhorn installed 
 - Port 8888 must be open and not used


## Install

For use apik3s software you can download the apik3s exec from desired release. \
You can use this by running it or create a new systemd service for automate api start on system boot for example:

```bash
mv apik3s /usr/local/bin/apik3s
vi /etc/systemd/system/apik3s.service
```

Insert:
```bash
[Unit]
Description=apik3s service

[Service]
ExecStart=/usr/local/bin/apik3s

[Install]
WantedBy=multi-user.target
```
next:

```bash
systemctl daemon-reload
systemctl enable --now apik3s.service
```

Now apik3s start listening on port 8888 automaticaly.
## How to use
For use this api a client must send http request to the api server. \
In Request sections you can see all supported request.

## GET Request

| request path | request scope |
| --- | --- |
| / | return a list of all supported request in a json file |
| /apik3s/storage/:namespace | return a list of active storage for a specific workspace/namespace |
| /apik3s/storage/ | return a list of active workspace/namespace |
| /apik3s/IPs | return a list of active ip address pool |


## POST Request

there are three post request for this api

| request path | requested body | request scope |
| --- | --- | --- |
| /apik3s/storage/root | {  <br>"id": "\*nfs storage name\*",  <br>"size": \*size in GB\*,  <br>"ip": "\*ip address or auto\*"  <br>} | used for create a new storage, it create a new pvc a new deploy of the storage with 1 pod replica and if it not exist a new namespace. |
| /apik3s/IPs | {  <br>"id": "\*pool name\*",  <br>"ips":\["\*pool1\*","\*pool2\*"\]  <br>} | it create a new ip pool, it create a new IP Address pool and a new L2 Advertisement, is possible specify ips with an ip interval or using IP CIDR |

## DELETE Request

there are three delete request in this api

| request path | request scope |
| --- | --- |
| /apik3s/storage/:namespace | it delete a workspace/namespace and all its content |
| /apik3s/storage/:namespace/:storage | it delete a single storage, delete his service, deploy and pvc |
| /apik3s/IPs/:poolname | it delete a ip pool, delete his ip address pool and his L2Advertisement |


## Used project

This api use different go packager:

- For api build: [gin-gonic](https://github.com/gin-gonic/gin)
- kubernetes client from: [kubernetes/client-go](https://github.com/kubernetes/client-go)
- metallb client from: [openconfig/kne](https://github.com/openconfig/kne/tree/main/api/metallb/clientset/v1beta1)
- metallb interfaces from: [metallb/metallb](https://github.com/metallb/metallb/tree/main/api/v1beta2)
- nfs container: [alphayax/docker-volume-nfs](https://github.com/alphayax/docker-volume-nfs)
## Authors

- [Massaroni Alex](https://www.github.com/rh363)
- [Vona Daniele](https://www.github.com/danielv99)

## Roadmap


- Make code more light

