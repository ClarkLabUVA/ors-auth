# Fairscape Authentication and Authorization Server

## Authorization Service

Authorization in Fairscape relies on a policy based model.
It can be loosley defined as a set of rules that under the correct *conditions*, define what actions the policy *principle* can take towards *resources*.

ALL microservices MUST for check permissions at the service with a *challenge*, a question checking whether the principle possesses the adequate rights to preform the request. This exists as a RESTfull request to the central auth service.

SOME microservices are responsbile for creating *resource* representations. These microservices are 

# Policies

Taking the most simple approach, we can just reduce the policies to read, write, delete

As groups and projects are also resources themselves that must be protected,
we apply those three policies to groups and projects as well.

# Definitions

### Principle
A principle represents the "who", the agent preforming resource requests of Fairscape. This can represent Users or Groups in our system.
### Resources
Resources represent the "what", the entity within the system that has many actions a principle may take. An example of resource in our system is a Dataset, with an Identifier, and a File in the object service.
### Actions
### Policy
### Challenge
### Group
A group is a set of users which can be entitled to a set of rights determined through policy.
### Project
A project is a set of resources which can be used to represent the set in policy. Users can create policy and add resources through the auth service. Additionaly users may specify the project on resource creation, in which case the endpoints must add resources to projects via RESTfull calls to the auth service.


