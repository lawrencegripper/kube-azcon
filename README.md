Very early spike to investigate allowing kubernetes cluster to operation on Azure PAAS services such as AzureSQL. 

Focusing on Postgresql paas service in first version. 

Aim:
- Install Customer Resource Definition defining AzureResource in kubectl 
- Create a yaml file defining postgres server name, username and password
- Custom controller to pick this up create the paas service. Create a service in Kubernetes so cluster services can resolve with clusterdns. Add secrets to cluster for password and username. 

Build

Install go dep 
dep restore
go build

Cluster install

todo: helm