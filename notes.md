# TODO List

Models
- User
  - ListAccess
  - ListOwned
  - ListPolicies
  - ListChallenges

Endpoints
- Error Handling:
  - ListUserHandler
  - GetUserHandler
  - DeleteUserHandler

handleErrors functions

New Endpoints

- Resource
  - Create
  - Get
  - Delete
  - List

- Policy
  - Create
  - Get
  - Update
  - Delete
  - List


- Challenge
  - Create
  - List

- Group
  - Create
  - Get
  - List
  - Update
  - Delete

# Auth Service Design Document



# Models

## Policy

### POST /policy

```
{
  "Resource": "ark:99999/test",
  "Principal": ["orcid:1234-1234-1234-1234", "group:12345", "*"],
  "Effect": "Allow",
  "Action": ["s3:GetObject", ..., "mds:*"],
  "Issuer": "<identifier for service>"
}
```

### GET /policy


### DELETE /policy/<policyId>


## Challenge

### POST /challenge


```
[
  {
    "Principal": "orcid:1234-1234-1234-1234",
    "Action": "S3:Download",
    "Object": "ark:99999/test"
  },
  ...
]
```

- Response

success is 200
failure is 403

```
{
  "timestamp": ...,
  "authorized": true,
}
```

#### Query for policies

```
bson.D{{
  "$and",
    bson.D{
      {"effect", "Allow"},
      {"object", "ark:99999/test"},
      {"action", ""},
      {"principal",
        bson.D{ "$eq",
          {}
        }
      }
    }
  }}
```

Principal may be
  - user
  - one of users groups
  - or * everyone

Action may be
  - Challenge.Action
  - or *

## Identity Based Policies

## Resource Based Policies

```
{
  "@type": "Policy"
  "Action": ["s3:GetObject"]
  "Principal": ["*"],
  "Effect": "Allow",
  "Resource": "ark:99999/sample"
}
```

Action -> For Service and ServiceResource
Resource -> Target of the Policy, always an Identifier
Effect -> Allow or Deny
Action -> List of Actions


Mongo Query For Finding A Policy by Name

```
{$find,
  { $and: {
      {"@type": {$eq: "Policy"}},
      {"Effect": {$eq: "Allow"}},
      {"Action": {$in: "s3:GetObject"}},
      {"Resource": {$eq: "ark:99999/sample"}},
    }
  }
}
```



## Actions By Service

### S3

resource s3:Object
  s3:GetObject
  s3:UpdateObject
  s3:CopyObject

resource s3:Bucket
  s3:ListObjects
  s3:CreateObject
  s3:DeleteBucket

Similarly Define Policies for HDFS

#### Identifier Service

resource mds:identifier

operations
  mds:GetIdentifier
  mds:UpdateIdentifier
  mds:DeleteIdentifier

resource mds:Namespace  

operations
  mds:CreateIdentifier
  mds:GetNamespace
  mds:CreateNamespace
  mds:DeleteNamespace



# Business logic for policy execution


# API Service Design

POST /policy
  one resource per policy

GET /policy?user=<userID>&resource=<identifier>&action=<action>
  list & filter policies

POST /challenge
  question does this user have this policy


### Dealing with Conflicting Policies

First Post a Policy that Grants Permissions
Post a Policy that Denies Permissions to same group

Find All Policies with overlapping Resource and a overlapping principal

```
{ $find: {
    $and: {
      {"Resource": <resource>},
      {"Effect": <opposite>},
      {"Action": {$in: [<actions of policy one>] }}
      {"Principal": {$elemMatch: [<principals of policy one>]}}
    }
  }
}
```

If it returns something, return an error


## General

### Errors to handle from Mongo

- mongo.ErrNoDocuments ("mongo: no documents in result")
