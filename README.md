# KloudFormation

KloudFormation is a Kubernetes operator (controller-manager) responsible for managing cloud infrastructure resources. Today it is a translation of selected AWS CloudFormation resources into Kubernetes resources. It uses custom resource definitions (CRDs) and the Kubebuilder scaffolding. However, unlike AWS CloudFormation, this project provides continuous state reconciliation for cloud infrastructure resources. 

This project also includes other higher-level abstractions. Those provide deployment and management of common infrastructure patterns such as three-tier networks, governance models, and clustered workload platforms.

## Status

This project is exiting the proof-of-concept phase and funded development is driving code refinement and resource catalog expansion. Testing is currently disabled as SIG-testing can't seem to provide a framework without side effects and grotesque leaks.

## Philosophy

OSS will eat the Cloud. But until it does we're going to find ways to eat at the value-add services they provide.

## AWS Resources

Resource selection has been based requirements for higher-level abstractions. Our initial target was launching a Docker Swarm cluster in AWS without any pre-existing environment. To accomplish that task we modeled many EC2 resources (VPC, subnets, gateways, and instances) and typical IAM resources. 

## KloudFormation Infrastructure Resources

### AuthorizeEC2SecurityGroupIngress

#### Description

AuthorizeEC2SecurityGroupIngress (soon to be renamed... too long) creates an ingress rule for an AWS EC2 Security Group

#### Spec Fields:

```
  - ruleName # string- k8s name of the rule. Used in the finalizer applied to the security group.
  - sourceCidrIp # string- ex. "0.0.0.0/0"
  - ec2SecurityGroupName # string- k8s name of the EC2SecurityGroup to assign the rule to.
  - fromPort # integer
  - toPort # integer
  - ipProtocol # string- tcp, udp, icmp, or protocol number. -1 is all protocols
```

#### Dependencies:

  - EC2SecurityGroup

### EC2Instance

#### Description

An EC2 instance. Launches 1 instance.

#### Spec Fields:

```
  - imageId # string- AMI number.
  - instanceType # string- ex. "t2.micro"
  - subnetName # string- k8s name of the subnet to be launched in
  - userData # string- (Use plaintext) Will be base64 encoded by the controller
  - ec2KeyPair # string- k8s name of the EC2 Key pair to use with the EC2 instance
  - ec2SecurityGroupName # string- k8s name of the EC2 security group to assign to the instance. Limit 1 for now.
  - tags
```

#### Dependencies:

  - Subnet
  - EC2KeyPair
  - EC2SecurityGroup

### EC2KeyPair
#### Description
An EC2 Keypair
#### Spec Fields:
```
  - ec2KeyPairName # string- AWS name of the keypair to create
```
#### Dependencies:

### EC2SecurityGroup
#### Description
  Creates an AWS EC2 Security Group.
#### Spec Fields:
```
  - ec2SecurityGroupName # string- AWS name for the security group.
  - vpcName # string- k8s name of the AWS VPC to place the security group in.
  - description # string-
  - tags
```
#### Dependencies:
  - EC2SecurityGroup
  - VPC

### EC2VolumeAttachment
#### Description
  Attaches an EBS Volume to an EC2 Instance.
#### Spec Fields:
```
  - devicePath # string
  - volumeName # string- k8s name of the AWS Volume
  - ec2InstanceName # string- k8s of the AWS EC2 instance to attach to
```  
#### Dependencies:
  - Volume
  - EC2Instance

### EIP
#### Description
  Creates an AWS Elastic IP
#### Spec Fields:
```
  - vpcName # string- k8s name of the AWS VPC to assign the EIP to.
  - tags
```  
#### Dependencies:
  - VPC

### EIPAssociation
#### Description
  Associates an EIP with an EC2 Instance.
#### Spec Fields:
```
  - allocationName # string- k8s name of an EIP to assign to an EC2 instance
  - ec2InstanceName # string- k8s name of the EC2 instance to assign the EIP to
```  
#### Dependencies:
  - EIP
  - EC2Instance

### InternetGateway
#### Description
  Creates an AWS Internet Gateway
#### Spec Fields:
```
  - vpcName # string- Unused. Need to remove.
  - tags
```  
#### Dependencies:
  - VPC

### InternetGatewayAttachment
#### Description
#### Spec Fields:
```
  - vpcName # string- k8s name of the VPC to attach the Internet Gateway to
  - internetGatewayName # string- the k8s name of the InternetGateway to attach to the VPC
```
#### Dependencies:
  - VPC
  - InternetGateway

### NATGateway
#### Description
#### Spec Fields:
```
  - subnetName # string- The k8s name of the Subnet to attach the NAT Gateway to
  - eipAllocationName # string- the k8s name of the EIP to use with the NAT Gateway
  - tags
```  
#### Dependencies:
  - Subnet
  - EIP

### Route
#### Description
  Route to an InternetGateway
#### Spec Fields:
```
  - destinationCidrBlock # string- the destination for the route. ex. "0.0.0.0/0"
  - routeTableName # string- the k8s name of the route table to assign the route to
  - gatewayName # string- the k8s name of the InternetGateway to use with the route
```  
#### Dependencies:
  - RouteTable
  - InternetGateway

### RouteTable
#### Description
#### Spec Fields:
```
  - vpcName # string- the k8s name of the VPC to create the route table within
  - tags
```  
#### Dependencies:
  - VPC

### RouteTableAssociation
#### Description
#### Spec Fields:
```
  - subnetName # string- the k8s name of the subnet to associate the Route Table with
  - routeTableName # string- the k8s name of the RouteTable being associated with the subnet
```  
#### Dependencies:
  - Subnet
  - RouteTable

### Subnet
#### Description
#### Spec Fields:
```
  - vpcName # string- the k8s name of the VPC to assign the Subnet to.
  - availabilityZone # string- the AWS availability zone to place the Subnet in.
  - cidrBlock # string- the CIDR range for the subnet. ex. "10.1.0.0/16"
  - tags
```  
#### Dependencies:
  - VPC

### Volume
#### Description
#### Spec Fields:
```
  - availabilityZone # string- the AWS availability zone to place the Volume in
  - size # int64- the size (in GB) of the Volume
  - volumeType # string- the type of Volume. "gp2", "io1", "st1", and "sc1" are valid values.
  - tags
```  
#### Dependencies:

### VPC
#### Description
#### Spec Fields:
```
  - cidrBlock # string- CIDR range of the VPC, ex. "10.0.0.0/8"
  - enableDnsSupport # string
  - enableDnsHostnames # string
  - instanceTenancy # string
  - tags
```  
#### Dependencies:

### AddRoleToInstanceProfile
#### Description
#### Spec Fields:
```
  - iamInstanceProfileName # string- k8s name of the AWS Instance Profile to add the AWS Role to
  - iamRoleName # string- k8s name of the AWS Role to add to the AWS Instance Profile
```  
#### Dependencies:
  - IAMInstanceProfile
  - Role

### IAMAttachRolePolicy
#### Description
#### Spec Fields:
```
  - iamPolicyName # string- the k8s name of the AWS IAM Policy to attach to the AWS Role.
  - iamRoleName # string- the k8s name of the AWS IAM Role to which the AWS IAM Policy will be added.
```  
#### Dependencies:
  - IAMPolicy
  - Role

### IAMInstanceProfile
#### Description
#### Spec Fields:
```
  - iamInstanceProfileName # string- the AWS name of the Instance Profile to create
  - path # string- the path of the Instance Profile
```  
#### Dependencies:

### IAMPolicy
#### Description
#### Spec Fields:
```
  - description # string
  - path # string- the path of the IAM Policy
  - policyDocument # string- the JSON policy document that defines the policy
  - policyName # string- the name to assign to the AWS IAM Policy
```  
#### Dependencies:

### Role
#### Description
#### Spec Fields:
```
  - assumeRolePolicyDocument # string- the JSON assume role policy document
  - description # string
  - maxSessionDuration # int64- maximum role session duration, in seconds
  - path # string- path for Role
  - roleName - # string- AWS name for the Role.
```  
#### Dependencies:

## KloudFormation Advanced Resources

### DockerSwarm
#### Description
#### Spec Fields:
```
  - numManagers
  - numWorkers
  - managerSize
  - workerSize
```  
#### Dependencies:

## Getting Started with Basic Resources

## Getting Started with Advanced Resources

## Notes
  - Need to make all names in CRDs specify aws or k8s for clarity. ex. awsKeyPairName or k8sKeyPairName.
  - Need to significantly shorten some names, and just change others for clarity and ease of use.
