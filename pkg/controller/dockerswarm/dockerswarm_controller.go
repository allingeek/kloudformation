/*
Copyright 2018 The KloudFormation authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dockerswarm

import (
	"context"
	//"fmt"
	//"log"
	"reflect"

	eccv1alpha1 "github.com/gotopple/kloudformation/pkg/apis/ecc/v1alpha1"
	swarmv1alpha1 "github.com/gotopple/kloudformation/pkg/apis/swarm/v1alpha1"
	//testv1alpha1 "github.com/gotopple/kloudformation/pkg/apis/test/v1alpha1"
	//iamv1alpha1 "github.com/gotopple/kloudformation/pkg/apis/iam/v1alpha1"
	//appsv1 "k8s.io/api/apps/v1"
	//corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new DockerSwarm Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this dockerswarm.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	r := mgr.GetRecorder(`dockerswarm-controller`)
	return &ReconcileDockerSwarm{Client: mgr.GetClient(), scheme: mgr.GetScheme(), events: r}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("dockerswarm-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to DockerSwarm
	err = c.Watch(&source.Kind{Type: &swarmv1alpha1.DockerSwarm{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a VPC created by DockerSwarm - change this for objects you create
	err = c.Watch(&source.Kind{Type: &eccv1alpha1.VPC{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &eccv1alpha1.VPC{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDockerSwarm{}

// ReconcileDockerSwarm reconciles a DockerSwarm object
type ReconcileDockerSwarm struct {
	client.Client
	scheme *runtime.Scheme
	events record.EventRecorder
}

// Reconcile reads that state of the cluster for a DockerSwarm object and makes changes based on the state read
// and what is in the DockerSwarm.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// Automatically generate RBAC rules to allow the Controller to read and write VPCs
// +kubebuilder:rbac:groups=ecc.aws.gotopple.com,resources=vpcs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=swarm.aws.gotopple.com,resources=dockerswarms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ecc.aws.gotopple.com,resources=subnets,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileDockerSwarm) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the DockerSwarm instance
	instance := &swarmv1alpha1.DockerSwarm{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// TODO(user): Change this to be the object type created by your controller
	// Define the desired VPC object
	r.events.Eventf(instance, `Normal`, `Info`, "Defining swarm VPC")
	vpc := &eccv1alpha1.VPC{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-vpc",
			Namespace: instance.Namespace,
		},
		Spec: eccv1alpha1.VPCSpec{
			CIDRBlock:          "10.20.0.0/16",
			EnableDNSSupport:   true,
			EnableDNSHostnames: true,
			InstanceTenancy:    "default",
			Tags: []eccv1alpha1.ResourceTag{
				{
					Key:   "Name",
					Value: "SwarmVPC",
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(instance, vpc, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Check if the VPC already exists
	r.events.Eventf(instance, `Normal`, `Info`, "Checking if swarm VPC exists")
	vpcFound := &eccv1alpha1.VPC{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: vpc.Name, Namespace: vpc.Namespace}, vpcFound)

	if err != nil && errors.IsNotFound(err) {
		r.events.Eventf(instance, `Normal`, `Info`, "Creating swarm VPC")
		err = r.Create(context.TODO(), vpc)
		if err != nil {
			r.events.Eventf(instance, `Normal`, `Info`, "Error in creating swarm VPC %s", err.Error())
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	r.events.Eventf(instance, `Normal`, `Info`, "Swarm VPC is present")
	err = r.Update(context.TODO(), vpc)

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(vpc.Spec, vpcFound.Spec) {
		vpcFound.Spec = vpc.Spec
		r.events.Eventf(instance, `Normal`, `Info`, "Updating swarm VPC")
		err = r.Update(context.TODO(), vpcFound)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	/// SUBNET STUFF ////
	// TODO(user): Change this to be the object type created by your controller
	// Define the desired VPC object
	r.events.Eventf(instance, `Normal`, `Info`, "Defining swarm VPC")
	subnet := &eccv1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-subnet",
			Namespace: instance.Namespace,
		},
		Spec: eccv1alpha1.SubnetSpec{
			VPCName:          vpc.Name,
			AvailabilityZone: "us-west-2a",
			CIDRBlock:        "10.20.110.0/24",
			Tags: []eccv1alpha1.ResourceTag{
				{
					Key:   "Name",
					Value: "SwarmSubnet",
				},
				{
					Key:   "SwarmOwner",
					Value: instance.Name,
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(instance, subnet, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Check if the VPC already exists
	r.events.Eventf(instance, `Normal`, `Info`, "Checking if swarm Subnet exists")
	subnetFound := &eccv1alpha1.Subnet{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: subnet.Name, Namespace: subnet.Namespace}, subnetFound)

	if err != nil && errors.IsNotFound(err) {
		r.events.Eventf(instance, `Normal`, `Info`, "Creating swarm Subnet")
		err = r.Create(context.TODO(), subnet)
		if err != nil {
			r.events.Eventf(instance, `Normal`, `Info`, "Error in creating swarm Subnet %s", err.Error())
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	r.events.Eventf(instance, `Normal`, `Info`, "Swarm Subnet is present")
	err = r.Update(context.TODO(), subnet)

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(subnet.Spec, subnetFound.Spec) {
		subnetFound.Spec = subnet.Spec
		r.events.Eventf(instance, `Normal`, `Info`, "Updating swarm Subnet")
		err = r.Update(context.TODO(), subnetFound)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	/// END SUBNET STUFF ///

	/// internetGatewayEIP STUFF ////
	// TODO(user): Change this to be the object type created by your controller
	// Define the desired VPC object
	r.events.Eventf(instance, `Normal`, `Info`, "Defining swarm InternetGatewayEIP")
	internetgatewayeip := &eccv1alpha1.EIP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-internetgateway-eip",
			Namespace: instance.Namespace,
		},
		Spec: eccv1alpha1.EIPSpec{
			VpcName: vpc.Name,
			Tags: []eccv1alpha1.ResourceTag{
				{
					Key:   "Name",
					Value: "SwarmInternetGatewayEIP",
				},
				{
					Key:   "SwarmOwner",
					Value: instance.Name,
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(instance, internetgatewayeip, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Check if the VPC already exists
	r.events.Eventf(instance, `Normal`, `Info`, "Checking if swarm InternetGatewayEIP exists")
	internetgatewayeipFound := &eccv1alpha1.EIP{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: internetgatewayeip.Name, Namespace: internetgatewayeip.Namespace}, internetgatewayeipFound)

	if err != nil && errors.IsNotFound(err) {
		r.events.Eventf(instance, `Normal`, `Info`, "Creating swarm InternetGatewayEIP")
		err = r.Create(context.TODO(), internetgatewayeip)
		if err != nil {
			r.events.Eventf(instance, `Normal`, `Info`, "Error in creating swarm InternetGatewayEIP %s", err.Error())
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	r.events.Eventf(instance, `Normal`, `Info`, "Swarm InternetGatewayEIP is present")
	err = r.Update(context.TODO(), internetgatewayeip)

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(internetgatewayeip.Spec, internetgatewayeipFound.Spec) {
		internetgatewayeipFound.Spec = internetgatewayeip.Spec
		r.events.Eventf(instance, `Normal`, `Info`, "Updating swarm InternetGatewayEIP")
		err = r.Update(context.TODO(), internetgatewayeipFound)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	/// END EIP STUFF ///

	/// INTERNETGATEWAY STUFF ////
	// TODO(user): Change this to be the object type created by your controller
	// Define the desired VPC object
	r.events.Eventf(instance, `Normal`, `Info`, "Defining swarm InternetGateway")
	internetGateway := &eccv1alpha1.InternetGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-internetgateway",
			Namespace: instance.Namespace,
		},
		Spec: eccv1alpha1.InternetGatewaySpec{
			VPCName: vpc.Name,
			Tags: []eccv1alpha1.ResourceTag{
				{
					Key:   "Name",
					Value: "SwarmInternetGateway",
				},
				{
					Key:   "SwarmOwner",
					Value: instance.Name,
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(instance, internetGateway, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Check if the VPC already exists
	r.events.Eventf(instance, `Normal`, `Info`, "Checking if swarm InternetGateway exists")
	internetGatewayFound := &eccv1alpha1.InternetGateway{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: internetGateway.Name, Namespace: internetGateway.Namespace}, internetGatewayFound)

	if err != nil && errors.IsNotFound(err) {
		r.events.Eventf(instance, `Normal`, `Info`, "Creating swarm InternetGateway")
		err = r.Create(context.TODO(), internetGateway)
		if err != nil {
			r.events.Eventf(instance, `Normal`, `Info`, "Error in creating swarm InternetGateway %s", err.Error())
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	r.events.Eventf(instance, `Normal`, `Info`, "Swarm InternetGateway is present")
	err = r.Update(context.TODO(), internetGateway)

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(internetGateway.Spec, internetGatewayFound.Spec) {
		internetGatewayFound.Spec = internetGateway.Spec
		r.events.Eventf(instance, `Normal`, `Info`, "Updating swarm InternetGateway")
		err = r.Update(context.TODO(), internetGatewayFound)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	/// END INTERNETGATEWAY STUFF ///

	/// INTERNETGATEWAYATTACHMENT STUFF ////
	// TODO(user): Change this to be the object type created by your controller
	// Define the desired VPC object
	r.events.Eventf(instance, `Normal`, `Info`, "Defining swarm InternetGatewayAttachment")
	internetGatewayAttachment := &eccv1alpha1.InternetGatewayAttachment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-internetgatewayattachment",
			Namespace: instance.Namespace,
		},
		Spec: eccv1alpha1.InternetGatewayAttachmentSpec{
			VPCName:             vpc.Name,
			InternetGatewayName: internetGateway.Name,
		},
	}
	if err := controllerutil.SetControllerReference(instance, internetGatewayAttachment, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Check if the VPC already exists
	r.events.Eventf(instance, `Normal`, `Info`, "Checking if swarm InternetGatewayAttachment exists")
	internetGatewayAttachmentFound := &eccv1alpha1.InternetGatewayAttachment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: internetGatewayAttachment.Name, Namespace: internetGatewayAttachment.Namespace}, internetGatewayAttachmentFound)

	if err != nil && errors.IsNotFound(err) {
		r.events.Eventf(instance, `Normal`, `Info`, "Creating swarm InternetGatewayAttachment")
		err = r.Create(context.TODO(), internetGatewayAttachment)
		if err != nil {
			r.events.Eventf(instance, `Normal`, `Info`, "Error in creating swarm InternetGatewayAttachment %s", err.Error())
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	r.events.Eventf(instance, `Normal`, `Info`, "Swarm InternetGatewayAttachment is present")
	err = r.Update(context.TODO(), internetGatewayAttachment)

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(internetGatewayAttachment.Spec, internetGatewayAttachmentFound.Spec) {
		internetGatewayAttachmentFound.Spec = internetGatewayAttachment.Spec
		r.events.Eventf(instance, `Normal`, `Info`, "Updating swarm InternetGatewayAttachment")
		err = r.Update(context.TODO(), internetGatewayAttachmentFound)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	/// END INTERNETGATEWAYATTACHMENT STUFF ///
	/// NATEIP STUFF ////
	// TODO(user): Change this to be the object type created by your controller
	// Define the desired VPC object
	r.events.Eventf(instance, `Normal`, `Info`, "Defining swarm SwarmNATGatewayEIP")
	natGatewayEIP := &eccv1alpha1.EIP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-natgateway-eip",
			Namespace: instance.Namespace,
		},
		Spec: eccv1alpha1.EIPSpec{
			VpcName: vpc.Name,
			Tags: []eccv1alpha1.ResourceTag{
				{
					Key:   "Name",
					Value: "SwarmNATGatewayEIP",
				},
				{
					Key:   "SwarmOwner",
					Value: instance.Name,
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(instance, natGatewayEIP, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	r.events.Eventf(instance, `Normal`, `Info`, "Checking if swarm NATGatewayEIP exists")
	natGatewayEIPFound := &eccv1alpha1.EIP{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: natGatewayEIP.Name, Namespace: natGatewayEIP.Namespace}, natGatewayEIPFound)

	if err != nil && errors.IsNotFound(err) {
		r.events.Eventf(instance, `Normal`, `Info`, "Creating swarm InternetGatewayAttachment")
		err = r.Create(context.TODO(), natGatewayEIP)
		if err != nil {
			r.events.Eventf(instance, `Normal`, `Info`, "Error in creating swarm NATGatewayEIP %s", err.Error())
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	r.events.Eventf(instance, `Normal`, `Info`, "Swarm NATGatewayEIP is present")
	err = r.Update(context.TODO(), natGatewayEIP)

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(natGatewayEIP.Spec, natGatewayEIPFound.Spec) {
		natGatewayEIPFound.Spec = natGatewayEIP.Spec
		r.events.Eventf(instance, `Normal`, `Info`, "Updating swarm NATGatewayEIP")
		err = r.Update(context.TODO(), natGatewayEIPFound)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	/// END NATEIP STUFF ///

	/// NATGATEWAYATTACHMENT STUFF ////
	// TODO(user): Change this to be the object type created by your controller
	r.events.Eventf(instance, `Normal`, `Info`, "Defining swarm NATGateway")

	natGateway := &eccv1alpha1.NATGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-natgateway",
			Namespace: instance.Namespace,
		},
		Spec: eccv1alpha1.NATGatewaySpec{
			SubnetName:        subnet.Name,
			EIPAllocationName: natGatewayEIP.Name,
			Tags: []eccv1alpha1.ResourceTag{
				{
					Key:   "Name",
					Value: "SwarmNATGateway",
				},
				{
					Key:   "SwarmOwner",
					Value: instance.Name,
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(instance, natGateway, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	r.events.Eventf(instance, `Normal`, `Info`, "Checking if swarm NATGateway exists")
	natGatewayFound := &eccv1alpha1.NATGateway{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: natGateway.Name, Namespace: natGateway.Namespace}, natGatewayFound)

	if err != nil && errors.IsNotFound(err) {
		r.events.Eventf(instance, `Normal`, `Info`, "Creating swarm NATGateway")
		err = r.Create(context.TODO(), natGateway)
		if err != nil {
			r.events.Eventf(instance, `Normal`, `Info`, "Error in creating swarm NATGateway %s", err.Error())
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	r.events.Eventf(instance, `Normal`, `Info`, "Swarm NATGateway is present")
	err = r.Update(context.TODO(), natGateway)

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(natGateway.Spec, natGatewayFound.Spec) {
		natGatewayFound.Spec = natGateway.Spec
		r.events.Eventf(instance, `Normal`, `Info`, "Updating swarm NATGateway")
		err = r.Update(context.TODO(), natGatewayFound)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	/// END NATGATEWAY STUFF ///

	/// ROUTE TABLE STUFF ////
	// TODO(user): Change this to be the object type created by your controller
	r.events.Eventf(instance, `Normal`, `Info`, "Defining swarm RouteTable")

	routeTable := &eccv1alpha1.RouteTable{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-routetable",
			Namespace: instance.Namespace,
		},
		Spec: eccv1alpha1.RouteTableSpec{
			VpcName: vpc.Name,
			Tags: []eccv1alpha1.ResourceTag{
				{
					Key:   "Name",
					Value: "SwarmRouteTable",
				},
				{
					Key:   "SwarmOwner",
					Value: instance.Name,
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(instance, routeTable, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	r.events.Eventf(instance, `Normal`, `Info`, "Checking if swarm RouteTable exists")
	routeTableFound := &eccv1alpha1.RouteTable{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: routeTable.Name, Namespace: routeTable.Namespace}, routeTableFound)

	if err != nil && errors.IsNotFound(err) {
		r.events.Eventf(instance, `Normal`, `Info`, "Creating swarm RouteTable")
		err = r.Create(context.TODO(), routeTable)
		if err != nil {
			r.events.Eventf(instance, `Normal`, `Info`, "Error in creating swarm RouteTable %s", err.Error())
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	r.events.Eventf(instance, `Normal`, `Info`, "Swarm RouteTable is present")
	err = r.Update(context.TODO(), routeTable)

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(routeTable.Spec, routeTableFound.Spec) {
		routeTableFound.Spec = routeTable.Spec
		r.events.Eventf(instance, `Normal`, `Info`, "Updating swarm RouteTable")
		err = r.Update(context.TODO(), routeTableFound)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	/// END ROUTE TABLE STUFF ///

	/// ROUTE STUFF ////
	// TODO(user): Change this to be the object type created by your controller
	r.events.Eventf(instance, `Normal`, `Info`, "Defining swarm Route")

	route := &eccv1alpha1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-route",
			Namespace: instance.Namespace,
		},
		Spec: eccv1alpha1.RouteSpec{
			DestinationCidrBlock: "0.0.0.0/0",
			GatewayName:          internetGateway.Name,
			RouteTableName:       routeTable.Name,
		},
	}
	if err := controllerutil.SetControllerReference(instance, route, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	r.events.Eventf(instance, `Normal`, `Info`, "Checking if swarm Route exists")
	routeFound := &eccv1alpha1.Route{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, routeFound)

	if err != nil && errors.IsNotFound(err) {
		r.events.Eventf(instance, `Normal`, `Info`, "Creating swarm RouteTable")
		err = r.Create(context.TODO(), route)
		if err != nil {
			r.events.Eventf(instance, `Normal`, `Info`, "Error in creating swarm Route %s", err.Error())
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	r.events.Eventf(instance, `Normal`, `Info`, "Swarm Route is present")
	err = r.Update(context.TODO(), route)

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(route.Spec, routeFound.Spec) {
		routeFound.Spec = route.Spec
		r.events.Eventf(instance, `Normal`, `Info`, "Updating swarm Route")
		err = r.Update(context.TODO(), routeFound)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	/// END ROUTE STUFF ///

	/// ROUTE TABLE ASSOCIATION STUFF ////
	// TODO(user): Change this to be the object type created by your controller
	r.events.Eventf(instance, `Normal`, `Info`, "Defining swarm RouteTableAssociation")

	routeTableAssociation := &eccv1alpha1.RouteTableAssociation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-routetableassociation",
			Namespace: instance.Namespace,
		},
		Spec: eccv1alpha1.RouteTableAssociationSpec{
			RouteTableName: routeTable.Name,
			SubnetName:     subnet.Name,
		},
	}
	if err := controllerutil.SetControllerReference(instance, routeTableAssociation, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	r.events.Eventf(instance, `Normal`, `Info`, "Checking if swarm RouteTableAssociation exists")
	routeTableAssociationFound := &eccv1alpha1.RouteTableAssociation{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: routeTableAssociation.Name, Namespace: routeTableAssociation.Namespace}, routeTableAssociationFound)

	if err != nil && errors.IsNotFound(err) {
		r.events.Eventf(instance, `Normal`, `Info`, "Creating swarm RouteTableAssociation")
		err = r.Create(context.TODO(), routeTableAssociation)
		if err != nil {
			r.events.Eventf(instance, `Normal`, `Info`, "Error in creating swarm RouteTableAssociation %s", err.Error())
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	r.events.Eventf(instance, `Normal`, `Info`, "Swarm RouteTableAssociation is present")
	err = r.Update(context.TODO(), routeTableAssociation)

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(routeTableAssociation.Spec, routeTableAssociationFound.Spec) {
		routeTableAssociationFound.Spec = routeTableAssociation.Spec
		r.events.Eventf(instance, `Normal`, `Info`, "Updating swarm RouteTableAssociation")
		err = r.Update(context.TODO(), routeTableAssociationFound)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	/// END ROUTE TABLE ASSOCIATION STUFF ///

	///  EC2 SECURITY GROUP STUFF ////
	// TODO(user): Change this to be the object type created by your controller
	r.events.Eventf(instance, `Normal`, `Info`, "Defining swarm EC2EC2SecurityGroup")

	ec2SecurityGroup := &eccv1alpha1.EC2SecurityGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-ec2securitygroup",
			Namespace: instance.Namespace,
		},
		Spec: eccv1alpha1.EC2SecurityGroupSpec{
			EC2SecurityGroupName: instance.Name + "-ec2securitygroup",
			VPCName:              vpc.Name,
			Description:          "A fun and exciting security group for the swarm",
		},
	}
	if err := controllerutil.SetControllerReference(instance, ec2SecurityGroup, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	r.events.Eventf(instance, `Normal`, `Info`, "Checking if swarm EC2SecurityGroup exists")
	ec2SecurityGroupFound := &eccv1alpha1.EC2SecurityGroup{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: ec2SecurityGroup.Name, Namespace: ec2SecurityGroup.Namespace}, ec2SecurityGroupFound)

	if err != nil && errors.IsNotFound(err) {
		r.events.Eventf(instance, `Normal`, `Info`, "Creating swarm EC2SecurityGroup")
		err = r.Create(context.TODO(), ec2SecurityGroup)
		if err != nil {
			r.events.Eventf(instance, `Normal`, `Info`, "Error in creating swarm EC2SecurityGroup %s", err.Error())
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	r.events.Eventf(instance, `Normal`, `Info`, "Swarm EC2SecurityGroup is present")
	err = r.Update(context.TODO(), ec2SecurityGroup)

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(ec2SecurityGroup.Spec, ec2SecurityGroupFound.Spec) {
		ec2SecurityGroupFound.Spec = ec2SecurityGroup.Spec
		r.events.Eventf(instance, `Normal`, `Info`, "Updating swarm EC2SecurityGroup")
		err = r.Update(context.TODO(), ec2SecurityGroupFound)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	/// END EC2 SECURITY GROUP STUFF ///

	///  EC2 SECURITY GROUP INGRESS STUFF ////
	// TODO(user): Change this to be the object type created by your controller
	r.events.Eventf(instance, `Normal`, `Info`, "Defining swarm AuthorizeEC2SecurityGroupIngress")

	authorizeEC2SecurityGroupIngress := &eccv1alpha1.AuthorizeEC2SecurityGroupIngress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-authorizeec2securitygroupingress",
			Namespace: instance.Namespace,
		},
		Spec: eccv1alpha1.AuthorizeEC2SecurityGroupIngressSpec{
			SourceCidrIp:         "0.0.0.0/0",
			EC2SecurityGroupName: ec2SecurityGroup.Name,
			FromPort:             22,
			ToPort:               22,
			IpProtocol:           "tcp",
		},
	}
	if err := controllerutil.SetControllerReference(instance, authorizeEC2SecurityGroupIngress, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	r.events.Eventf(instance, `Normal`, `Info`, "Checking if swarm AuthorizeEC2SecurityGroupIngress exists")
	authorizeEC2SecurityGroupIngressFound := &eccv1alpha1.AuthorizeEC2SecurityGroupIngress{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: authorizeEC2SecurityGroupIngress.Name, Namespace: authorizeEC2SecurityGroupIngress.Namespace}, authorizeEC2SecurityGroupIngressFound)

	if err != nil && errors.IsNotFound(err) {
		r.events.Eventf(instance, `Normal`, `Info`, "Creating swarm AuthorizeEC2SecurityGroupIngress")
		err = r.Create(context.TODO(), authorizeEC2SecurityGroupIngress)
		if err != nil {
			r.events.Eventf(instance, `Normal`, `Info`, "Error in creating swarm AuthorizeEC2SecurityGroupIngress %s", err.Error())
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	r.events.Eventf(instance, `Normal`, `Info`, "Swarm AuthorizeEC2SecurityGroupIngress is present")
	err = r.Update(context.TODO(), authorizeEC2SecurityGroupIngress)

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(authorizeEC2SecurityGroupIngress.Spec, authorizeEC2SecurityGroupIngressFound.Spec) {
		authorizeEC2SecurityGroupIngressFound.Spec = authorizeEC2SecurityGroupIngress.Spec
		r.events.Eventf(instance, `Normal`, `Info`, "Updating swarm AuthorizeEC2SecurityGroupIngress")
		err = r.Update(context.TODO(), authorizeEC2SecurityGroupIngressFound)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	/// END EC2 SECURITY GROUP INGRESS STUFF ///

	return reconcile.Result{}, nil
}
