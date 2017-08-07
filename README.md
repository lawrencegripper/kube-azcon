Introduction
===========

*WARNING* This is a spike/work in-progress

When a service requires, for example, MSSQL, Postgres, BlobStorage or Redis they can be requested from Azure through the kubernetes toolchain.

The controller provisions the service and creates a service inside the cluster, along with secrets for login,
to allow apps inside the cluster to use the resource as if it was deployed in the cluster.

Details
--------

Very early spike to investigate allowing kubernetes cluster to operation on Azure PAAS and investigate.

Aim:
- Install Customer Resource Definition defining AzureResource in kubectl 
- Create a yaml file defining a type of resource, name and other configuration details
- Custom controller to create the paas service then create a service in Kubernetes so cluster services can resolve with clusterdns and add secrets to cluster for authentication

Build

install glide
glide up -v (strips nested vendor files)
go build

Cluster install

todo: helm