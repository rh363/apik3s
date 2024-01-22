/*
NAME=APIK3S
AUTHOR=RH363
DATE=12/2023
COMPANY=SEEWEB
VERSION=1.5

DESCRIPTION:

THIS IS AN API WRITE IN GO-LANG IT'S POURPOSE IS COMMUNICATE WITH AN K3S
CLUSTER WITH THIS PC AS MASTER NODE FOR MANAGE NFS SERVER INSTANCE AND THEIR SERVICE.
THIS FILE CAN RUN ONLY IF EVERY COMPONENT IS IN ITS PLACE.
PLEASE READ README DOC BEFORE DEPLOY THIS API.


REQUIREMENT:
	-1: RUN IT AS ROOT
	-2: K3S INSTALLED AND CONFIGURED HOW DESCRIPTED README.TXT
	-3: CAN USE PORT 2049,20048,111,8888 AND ALL K3S PORT
*/

package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	v1beta1 "github.com/openconfig/kne/api/metallb/clientset/v1beta1"
	metallbv1 "go.universe.tf/metallb/api/v1beta1"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
	v1 "k8s.io/client-go/applyconfigurations/meta/v1"
)

var clientset *kubernetes.Clientset            //client set for k3s api
var metallbClientSet *v1beta1.Clientset        //client set for metallb api
var ConfigMapName string = "nfs-server-conf"   //configmap name
var MetallbNamespace string = "metallb-system" //metallb default namespace

// // k3s api ERROR
// deploy ERROR
var ErrDeployAlreadyExist error = errors.New("deploy already exist")
var ErrCantCreateDeploy error = errors.New("cannot create deploy")
var ErrCantDeleteDeploy error = errors.New("cannot delete deploy")
var ErrCantGetDeploy error = errors.New("cannot get deploy")
var ErrContainerUnsupported = errors.New("actually this server type is not supported")

// service ERROR
var ErrServiceAlreadyExist error = errors.New("service already exist")
var ErrCantCreateService error = errors.New("cannot create Service")
var ErrCantDeleteService error = errors.New("cannot delete Service")
var ErrCantGetService error = errors.New("cannot get Service")

// pvc ERROR
var ErrPVCAlreadyExist error = errors.New("PVC already exist")
var ErrCantCreatePVC error = errors.New("cannot create PVC")
var ErrCantDeletePVC error = errors.New("cannot delete PVC")
var ErrCantGetPVC error = errors.New("cannot get PVC")
var ErrPVCNotFound error = errors.New("PVC not found")

// namespace ERROR
var ErrNamespaceAlreadyExist error = errors.New("namespace already exist")
var ErrCantGetNamespace error = errors.New("cannot get namespace")
var ErrCantCreateNamespace error = errors.New("cannot create namespace")
var ErrCantDeleteNamespace error = errors.New("cannot delete namespace")

// L2Advertisement error
var ErrCantGetL2Advertisement error = errors.New("cannot get L2Advertisement")
var ErrCantCreateL2Advertisement error = errors.New("cannot create L2Advertisement")
var ErrCantDeleteL2Advertisement error = errors.New("cannot delete L2Advertisement")
var ErrL2AdvertisementAlreadyExist error = errors.New("L2Advertisement already exist")

// ip address pool error
var ErrCantGetIPAddressPool error = errors.New("cannot get ip address pool")
var ErrCantCreateIPAddressPool error = errors.New("cannot create ip address pool")
var ErrCantDeleteIPAddressPool error = errors.New("cannot delete ip address pool")
var ErrIPAddressPoolAlreadyExist error = errors.New("IPAddressPool already exist")

// configmap ERROR
var ErrCantCreateConfigMap error = errors.New("cannot create configmap")
var ErrCantDeleteConfigMap error = errors.New("cannot delete configmap")
var ErrCantGetConfigmap error = errors.New("cannot get configmap")

// // apik3s ERROR
// JSON ERROR
var ErrBadJsonFormat error = errors.New("json format used is not valid")

// size ERROR
var ErrInvalidSize error = errors.New("storage size received is not valid,insert only int > 0")
var ErrSizeToSmall error = errors.New("storage size received is not valid insert only value > actual storage size")

// Config File Path
var ConfigFilePath string = "/etc/rancher/k3s/k3s.yaml"

// ------------------------------------------------------------------------------------------------------------------------------utility function
/*func prompt() { //it simple stop until user dont click enter, for test pourpouse
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println()
}
*/
func int32Ptr(i int32) *int32 { return &i } //change int format,used for replica

func setTrue() *bool { //return true value in specific format
	b := true
	return &b
}

func makeString(s string) *string { //return string in specific format
	return &s
}

// ------------------------------------------------------------------------------------------------------------------------------k3s api functions
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ipadresspool functions
// get ip address pool function, return ip address list and an error if present withouth printing list
func getIPAddressPoolsDEV() (*metallbv1.IPAddressPoolList, error) {

	fmt.Println("Listing IP address pool: ")
	ipAddressPoolClient := metallbClientSet.IPAddressPool(MetallbNamespace) //create ip address pool client for api

	list, err := ipAddressPoolClient.List(context.TODO(), metav1.ListOptions{}) //get ip address pool list
	if err != nil {
		fmt.Println(err)
		return nil, ErrCantGetIPAddressPool
	}
	return list, nil
}

/*
func getIPAddressPools() (*metallbv1.IPAddressPoolList, error) { //get ip address pool function,return ip address list and an error if present

		fmt.Println("Listing IP address pool: ")
		ipAddressPoolClient := metallbClientSet.IPAddressPool(MetallbNamespace) //create ip address pool client for api

		list, err := ipAddressPoolClient.List(context.TODO(), metav1.ListOptions{}) //get ip address pool list
		if err != nil {
			fmt.Println(err)
			return nil, ErrCantGetIPAddressPool
		}

		for _, ipaddresspool := range list.Items {
			fmt.Println("+ip address pool: " + ipaddresspool.Name)
			for i := 0; i < len(ipaddresspool.Spec.Addresses); i++ {
				fmt.Println("+--> ips: " + ipaddresspool.Spec.Addresses[i])
			}
		}
		return list, nil
	}
*/
func createIPAdressPool(poolname string, pool []string) error { // create a new ip address pool require ip address pool name and pool array
	fmt.Println("[0/3]try to create new namespace...")

	list, err := getnamespacesDEV() //get current ip address pools

	if err != nil {
		return ErrCantGetIPAddressPool
	}

	for _, ipaddresspool := range list.Items {
		if ipaddresspool.Name == poolname {
			fmt.Println("ip address pool: \"" + poolname + "\" already exists") //check if ipaddress already exist
			return ErrIPAddressPoolAlreadyExist
		}
	}

	ipaddresspoolclient := metallbClientSet.IPAddressPool(MetallbNamespace) //create ipaddress client for api
	fmt.Println("[1/3]client set acquired")

	ipaddresspool := &metallbv1.IPAddressPool{ //declare ipaddress
		ObjectMeta: metav1.ObjectMeta{
			Name:      poolname,
			Namespace: MetallbNamespace,
		},
		Spec: metallbv1.IPAddressPoolSpec{
			Addresses: pool,
		},
	}

	fmt.Println("[2/3]ip address pool declared")
	result, err := ipaddresspoolclient.Create(context.TODO(), ipaddresspool, metav1.CreateOptions{}) //create ip address pool
	if err != nil {
		fmt.Println(err)
		return ErrCantCreateIPAddressPool
	}
	fmt.Printf("[3/3]ip address pool created: \"%q\".\n", result.GetObjectMeta().GetName())
	return nil
}

func deleteIPAddressPool(id string) error { //delete ip address pool function(target id)
	fmt.Println("Deleting ip address pool...")
	ipaddresspoolclient := metallbClientSet.IPAddressPool(MetallbNamespace)        //create ipaddress client for api
	deletePolicy := metav1.DeletePropagationForeground                             //set delete policy
	if err := ipaddresspoolclient.Delete(context.TODO(), id, metav1.DeleteOptions{ //delete ip address pool
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
		return ErrCantDeleteIPAddressPool
	}
	fmt.Println("ip address pool deleted.")
	return nil
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ L2 Advertisement
func getL2AdvertisementDEV() (*metallbv1.L2AdvertisementList, error) { //get L2Advertisement function

	fmt.Println("Listing IP address pool: ")
	L2Advertisementclient := metallbClientSet.L2Advertisement(MetallbNamespace) //create L2Advertisement client for api

	list, err := L2Advertisementclient.List(context.TODO(), metav1.ListOptions{}) //get L2Advertisement list
	if err != nil {
		fmt.Println(err)
		return nil, ErrCantGetL2Advertisement
	}
	return list, nil
}

/*
func getL2Advertisement() (*metallbv1.L2AdvertisementList, error) { //get L2Advertisement function

		fmt.Println("Listing IP address pool: ")
		L2AdvertisementClient := metallbClientSet.L2Advertisement(MetallbNamespace) //create L2Advertisement client for api

		list, err := L2AdvertisementClient.List(context.TODO(), metav1.ListOptions{}) //get L2Advertisement list
		if err != nil {
			fmt.Println(err)
			return nil, ErrCantGetL2Advertisement
		}

		for _, advertisement := range list.Items {
			fmt.Println("+L2Advertisement: " + advertisement.Name)
			for i := 0; i < len(advertisement.Spec.IPAddressPools); i++ {
				fmt.Println("+--> ip address pool: " + advertisement.Spec.IPAddressPools[i])
			}
		}
		return list, nil
	}
*/
func createL2Advertisement(L2AdvertisementName string) error { // create L2Advertisement function
	fmt.Println("[0/3]try to create new L2Advertisement...")

	list, err := getL2AdvertisementDEV() //check if L2Advertisement already exist

	if err != nil {
		return ErrCantGetL2Advertisement
	}

	for _, advertisement := range list.Items {
		if advertisement.Name == L2AdvertisementName {
			fmt.Println("L2Advertisement: \"" + L2AdvertisementName + "\" already exists")
			return ErrNamespaceAlreadyExist
		}
	}

	L2Advertisementclient := metallbClientSet.L2Advertisement(MetallbNamespace) //create L2Advertisement for api
	fmt.Println("[1/3]client set acquired")

	advertisement := &metallbv1.L2Advertisement{
		ObjectMeta: metav1.ObjectMeta{
			Name:      L2AdvertisementName,
			Namespace: MetallbNamespace,
		},
		Spec: metallbv1.L2AdvertisementSpec{
			IPAddressPools: []string{
				L2AdvertisementName,
			},
		},
	}

	fmt.Println("[2/3]L2Advertisement declared")
	result, err := L2Advertisementclient.Create(context.TODO(), advertisement, metav1.CreateOptions{}) //create L2Advertisement
	if err != nil {
		fmt.Println(err)
		return ErrCantCreateL2Advertisement
	}
	fmt.Printf("[3/3]L2Advertisement created: \"%q\".\n", result.GetObjectMeta().GetName())
	return nil
}

func deleteL2Advertisement(id string) error { //delete L2Advertisement function(target id)
	fmt.Println("Deleting L2Advertisement...")
	L2Advertisementclient := metallbClientSet.L2Advertisement(MetallbNamespace)      //create L2Advertisement client for api
	deletePolicy := metav1.DeletePropagationForeground                               //set delete policy
	if err := L2Advertisementclient.Delete(context.TODO(), id, metav1.DeleteOptions{ //delete L2Advertisement
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
		return ErrCantDeleteL2Advertisement
	}
	fmt.Println("L2Advertisement deleted.")
	return nil
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ deploys function
func createNFSdeploy(namespace string, ID string, pvc string) error { //create nfs deploy function (namespace target,deploy name and pod name(use it for match label),pvc volume name)
	fmt.Println("[0/3]try to create new nfs deploy...")

	list, err := getdeploysDEV(namespace) //check if deploy already exist

	if err != nil {
		return ErrCantGetDeploy
	}

	for _, deploy := range list.Items {
		if deploy.Name == ID {
			fmt.Println("nfs deploy: \"" + ID + "\" already exists")
			return ErrDeployAlreadyExist
		}
	}

	deploymentsClient := clientset.AppsV1().Deployments(namespace) //create deployment client for api
	fmt.Println("[1/3]client set acquired")

	deployment := &appsv1.Deployment{ //create deployment for api (use similar json format)
		ObjectMeta: metav1.ObjectMeta{
			Name: ID, //deployment id
			Labels: map[string]string{
				"type": "nfs",
			},
		},
		Spec: appsv1.DeploymentSpec{ //deployment specs
			Replicas: int32Ptr(1), //replica number
			Selector: &metav1.LabelSelector{ //deployment selector
				MatchLabels: map[string]string{
					"ID": ID,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{ //pod label for selector
						"ID": ID,
					},
				},
				Spec: apiv1.PodSpec{ //pod specs
					Containers: []apiv1.Container{
						{
							Name:  "nfs-server",                 //container name
							Image: "alphayax/docker-volume-nfs", //container image
							Ports: []apiv1.ContainerPort{
								{
									Name:          "nfs", //nfs port
									ContainerPort: 2049,
								},
								{
									Name:          "mountd", //mountd port
									ContainerPort: 20048,
								},
								{
									Name:          "rpcbind", //rpcbin port
									ContainerPort: 111,
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Privileged: setTrue(), //set security context to privileged
							},
							VolumeMounts: []apiv1.VolumeMount{ //mount pvc volume
								{
									Name:      "storage",
									MountPath: "/exports",
								},
								{
									Name:      ConfigMapName,
									MountPath: "/etc/exports.d/",
								},
							},
						},
					},
					Volumes: []apiv1.Volume{ //declare pvc volume
						{
							Name: "storage",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvc,
								},
							},
						},
						{
							Name: ConfigMapName,
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: ConfigMapName,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	fmt.Println("[2/3]nfs deploy declared")
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{}) //create deploy using deployment client and deploy file
	if err != nil {
		fmt.Println(err)
		return ErrCantCreateDeploy
	}
	fmt.Printf("[3/3]Created nfs deployment \"%q\".\n", result.GetObjectMeta().GetName())
	return nil
}

func deletedeploy(namespace string, ID string) error { //delete deploy function (namspace target,deploy id to delete)
	fmt.Println("Deleting deployment...")
	deploymentsClient := clientset.AppsV1().Deployments(namespace)               //create deployment client for api
	deletePolicy := metav1.DeletePropagationForeground                           //set delete policy
	if err := deploymentsClient.Delete(context.TODO(), ID, metav1.DeleteOptions{ //delete deploy
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
		return ErrCantDeleteDeploy
	}
	fmt.Println("Deployment deleted.")
	return nil
}

/*
func getdeploys(namespace string) (*appsv1.DeploymentList, error) { //get deploy function(namespace targets)

		fmt.Println("Listing deployments in namespace: " + namespace)
		deploymentsClient := clientset.AppsV1().Deployments(namespace) //create deployment client for api

		list, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{}) //get deployments list
		if err != nil {
			fmt.Println(err)
			return nil, ErrCantGetDeploy
		}

		for _, deploy := range list.Items { //print deployments list
			fmt.Println("+deploy: \"" + deploy.Name + "\"")
		}

		return list, nil
	}
*/
func getdeploysDEV(namespace string) (*appsv1.DeploymentList, error) { //get deploy function for DEV pourpose,it return the list but dont print it(namespace targets) list return

	deploymentsClient := clientset.AppsV1().Deployments(namespace) //create deployment client for api

	list, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{}) //get deployments list
	if err != nil {
		fmt.Println(err)
		return nil, ErrCantGetDeploy
	}
	return list, nil //return deployments list
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++service functions
func createNFSservice(serviceID string, namespace string, ID string, IP string) error { //create nfs service function (service ID(service name),namespace target,deployment target id,external ip)
	fmt.Println("[0/3]try to create new nfs service...")

	list, err := getservicesDEV(namespace) //check if service already exist

	if err != nil {
		return ErrCantGetService
	}

	for _, service := range list.Items {
		if service.Name == serviceID {
			fmt.Println("nfs service: \"" + serviceID + "\" already exists")
			return ErrServiceAlreadyExist
		}
	}

	svcClient := clientset.CoreV1().Services(namespace) //create service client for api
	fmt.Println("[1/3]client set acquired")

	service := &apiv1.Service{ //create service file for k3s api
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceID,
			Annotations: map[string]string{
				"metallb.universe.tf/loadBalancerIPs": IP,
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeLoadBalancer,
			Ports: []apiv1.ServicePort{ //exposed ports
				{
					Name: "nfs",
					Port: 2049,
				},
				{
					Name: "mountd",
					Port: 20048,
				},
				{
					Name: "rpcbind",
					Port: 111,
				},
			},
			Selector: map[string]string{ //deployments to expose, find by ID
				"ID": ID,
			},
		},
	}

	if IP == "auto" || IP == "" {
		service = &apiv1.Service{ //create service file for k3s api
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceID,
			},
			Spec: apiv1.ServiceSpec{
				Type: apiv1.ServiceTypeLoadBalancer,
				Ports: []apiv1.ServicePort{ //exposed ports
					{
						Name: "nfs",
						Port: 2049,
					},
					{
						Name: "mountd",
						Port: 20048,
					},
					{
						Name: "rpcbind",
						Port: 111,
					},
				},
				Selector: map[string]string{ //deployments to expose, find by ID
					"ID": ID,
				},
			},
		}
	}

	fmt.Println("[2/3]nfs service declared")
	result, err := svcClient.Create(context.TODO(), service, metav1.CreateOptions{}) //create service
	if err != nil {
		fmt.Println(err)
		return ErrCantCreateService
	}
	fmt.Printf("[3/3]nfs Service created: \"%q\".\n", result.GetObjectMeta().GetName())
	return nil
}

func deleteservice(namespace string, serviceID string) error { //delete service function (nampespace target,service ID to delete)
	fmt.Println("Deleting service...")
	svcClient := clientset.CoreV1().Services(namespace)                         //create service client for api
	deletePolicy := metav1.DeletePropagationForeground                          //set delete policy
	if err := svcClient.Delete(context.TODO(), serviceID, metav1.DeleteOptions{ //delete service
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
		return ErrCantDeleteService
	}
	fmt.Println("Service deleted.")
	return nil
}

/*
//return all active k3s service in a namespace
func getservices(namespace string) (*apiv1.ServiceList, error) { //get services function(namespace target)

		fmt.Println("Listing services in namespace: " + namespace)
		svcClient := clientset.CoreV1().Services(namespace) //create service client for api

		list, err := svcClient.List(context.TODO(), metav1.ListOptions{}) //get service list
		if err != nil {
			fmt.Println(err)
			return nil, ErrCantGetService
		}

		for _, svc := range list.Items { //for all service in service list print the service name and all service ip
			if svc.Name != "kubernetes" {
				fmt.Println("+service: " + svc.Name)
				for i := 0; i < len(svc.Status.LoadBalancer.Ingress); i++ {
					fmt.Println("+--->ip: " + svc.Status.LoadBalancer.Ingress[i].IP)
				}
			}
		}
		return list, nil
	}
*/
func getservicesDEV(namespace string) (*apiv1.ServiceList, error) { //get service function for DEV pourpose,only return list but dont print it(clientset,namespace target)

	svcClient := clientset.CoreV1().Services(namespace) //create service client for api

	list, err := svcClient.List(context.TODO(), metav1.ListOptions{}) //get service list
	if err != nil {
		fmt.Println(err)
		return nil, ErrCantGetService
	}
	return list, nil
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++pvc functions
func createpvc(namespace string, ID string, size string) error { //create pvc function (namespace target,pvc ID,size in string format(ex: 2Gi))
	fmt.Println("[0/3]try to create new pvc...")

	list, err := getpvcsDEV(namespace) //check if pvc already exist

	if err != nil {
		return ErrCantGetPVC
	}

	for _, pvc := range list.Items {
		if pvc.Name == ID {
			fmt.Println("pvc: \"" + ID + "\" already exists")
			return ErrPVCAlreadyExist
		}
	}

	pvcClient := clientset.CoreV1().PersistentVolumeClaims(namespace) //create pvc client for api
	fmt.Println("[1/3]client set acquired")

	persistentVolumeClaim := &apiv1.PersistentVolumeClaim{ //declare persistent volume claim
		ObjectMeta: metav1.ObjectMeta{
			Name: ID,
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				apiv1.ReadWriteOnce,
			},
			StorageClassName: makeString("longhorn"),
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceStorage: resource.MustParse(size),
				},
			},
		},
	}

	fmt.Println("[2/3]pvc declared")
	result, err := pvcClient.Create(context.TODO(), persistentVolumeClaim, metav1.CreateOptions{}) //create persistent volume claim
	if err != nil {
		fmt.Println(err)
		return ErrCantCreatePVC
	}
	fmt.Printf("[3/3]Persistent volume claim created: \"%q\".\n", result.GetObjectMeta().GetName())
	return nil
}

func updatepvc(namespace string, ID string, size string) error { //update pvc function (namespace target,pvc ID,size in string format(ex: 2Gi))
	fmt.Println("[0/3]try to create new pvc...")

	list, err := getpvcsDEV(namespace) //check if pvc exist

	if err != nil {
		return ErrCantGetPVC
	}

	for _, pvc := range list.Items {
		if pvc.Name == ID {
			fmt.Println("pvc found")
			pvcClient := clientset.CoreV1().PersistentVolumeClaims(namespace) //create pvc client for api
			fmt.Println("[1/3]client set acquired")

			persistentVolumeClaim := &corev1.PersistentVolumeClaimApplyConfiguration{
				TypeMetaApplyConfiguration: v1.TypeMetaApplyConfiguration{
					Kind:       makeString("PersistentVolumeClaim"),
					APIVersion: makeString("v1"),
				},
				ObjectMetaApplyConfiguration: &v1.ObjectMetaApplyConfiguration{
					Name: &ID,
				},
				Spec: &corev1.PersistentVolumeClaimSpecApplyConfiguration{
					AccessModes: []apiv1.PersistentVolumeAccessMode{
						apiv1.ReadWriteOnce,
					},
					StorageClassName: makeString("longhorn"),
					Resources: &corev1.ResourceRequirementsApplyConfiguration{
						Requests: &apiv1.ResourceList{
							apiv1.ResourceStorage: resource.MustParse(size),
						},
					},
				},
			}

			fmt.Println("[2/3]pvc declared")
			result, err := pvcClient.Apply(context.TODO(), persistentVolumeClaim, metav1.ApplyOptions{FieldManager: "api", Force: true}) //create persistent volume claim
			if err != nil {
				fmt.Println(err)
				return ErrCantCreatePVC
			}
			fmt.Printf("[3/3]Persistent volume claim updated: \"%q\".\n", result.GetObjectMeta().GetName())
			return nil
		}
	}
	return ErrPVCNotFound

}

func deletepvc(namespace string, ID string) error { //delete pvc function(namespace target,pvc id to delete)
	fmt.Println("Deleting pvc...")
	pvcClient := clientset.CoreV1().PersistentVolumeClaims(namespace)    //create pvc client for k3s api
	deletePolicy := metav1.DeletePropagationForeground                   //set delete policy
	if err := pvcClient.Delete(context.TODO(), ID, metav1.DeleteOptions{ //delete pvc
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
		return ErrCantDeletePVC
	}
	fmt.Println("Pvc deleted.")
	return nil
}

func getpvcs(namespace string) (*apiv1.PersistentVolumeClaimList, error) { //get pvcs function(namespace targets)

	fmt.Println("Listing pvc in namespace: " + namespace)
	pvcClient := clientset.CoreV1().PersistentVolumeClaims(namespace) //create pvc client for api

	list, err := pvcClient.List(context.TODO(), metav1.ListOptions{}) //get pvc list
	if err != nil {
		fmt.Println(err)
		return nil, ErrCantGetPVC
	}

	for _, pvc := range list.Items { //print pvc list
		fmt.Println("+pvc: " + pvc.Name + " " + pvc.Spec.Resources.Requests.Storage().String())
	}

	return list, nil
}

func getpvcsDEV(namespace string) (*apiv1.PersistentVolumeClaimList, error) { //get pvcs function for DEV pourpose,it only return list but dont print it(namesace targets)

	pvcClient := clientset.CoreV1().PersistentVolumeClaims(namespace) //create pvc client for k3s api

	list, err := pvcClient.List(context.TODO(), metav1.ListOptions{}) //get pvc list
	if err != nil {
		fmt.Println(err)
		return nil, ErrCantGetPVC
	}
	return list, nil //return pvc list
}

func getpvcsbyid(namespace string, id string) (*apiv1.PersistentVolumeClaim, error) { //get pvcs function(namespace targets)

	fmt.Println("Listing pvc in namespace: " + namespace)
	pvcClient := clientset.CoreV1().PersistentVolumeClaims(namespace) //create pvc client for api

	list, err := pvcClient.List(context.TODO(), metav1.ListOptions{}) //get pvc list
	if err != nil {
		fmt.Println(err)
		return nil, ErrCantGetPVC
	}

	for _, pvc := range list.Items { //print pvc list
		if pvc.Name == id {
			return &pvc, nil
		}
	}

	return nil, ErrPVCNotFound
}

// ------------------------------------------------------------------------------------------------------------------------------namespace
func createnamespace(Namespace string) error { //create namespace function (namespace name))
	fmt.Println("[0/3]try to create new namespace...")

	list, err := getnamespacesDEV() //check if namespace already exist

	if err != nil {
		return ErrCantGetNamespace
	}

	for _, namespace := range list.Items {
		if namespace.Name == Namespace {
			fmt.Println("namespace: \"" + Namespace + "\" already exists")
			return ErrNamespaceAlreadyExist
		}
	}

	namespaceClient := clientset.CoreV1().Namespaces() //create namespace client for api
	fmt.Println("[1/3]client set acquired")

	namespace := &apiv1.Namespace{ //declare namespace
		ObjectMeta: metav1.ObjectMeta{
			Name: Namespace,
		},
	}

	fmt.Println("[2/3]namespace declared")
	result, err := namespaceClient.Create(context.TODO(), namespace, metav1.CreateOptions{}) //create namespace
	if err != nil {
		fmt.Println(err)
		return ErrCantCreateNamespace
	}
	fmt.Printf("[3/3]namespace created: \"%q\".\n", result.GetObjectMeta().GetName())
	return nil
}

/*
func getnamespaces() (*apiv1.NamespaceList, error) { //get namespaces function()

		fmt.Println("Listing namespace: ")
		namespaceClient := clientset.CoreV1().Namespaces() //create namespace client for api

		list, err := namespaceClient.List(context.TODO(), metav1.ListOptions{}) //get namespace list
		if err != nil {
			fmt.Println(err)
			return nil, ErrCantGetNamespace
		}

		for _, namespace := range list.Items { //print namespace list
			fmt.Println("+namespace: " + namespace.Name)
		}

		return list, nil
	}
*/
func getnamespacesDEV() (*apiv1.NamespaceList, error) { //get namespaces function

	fmt.Println("Listing namespace: ")
	namespaceClient := clientset.CoreV1().Namespaces() //create namespace client for api

	list, err := namespaceClient.List(context.TODO(), metav1.ListOptions{}) //get namespace list
	if err != nil {
		fmt.Println(err)
		return nil, ErrCantGetNamespace
	}

	return list, nil
}

func deletenamespace(Namespace string) error { //delete namespace function(namespace target)
	fmt.Println("Deleting pvc...")
	namespaceClient := clientset.CoreV1().Namespaces()                                //create namespace client for k3s api
	deletePolicy := metav1.DeletePropagationForeground                                //set delete policy
	if err := namespaceClient.Delete(context.TODO(), Namespace, metav1.DeleteOptions{ //delete namespace
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
		return ErrCantDeleteNamespace
	}
	fmt.Println("Namespace deleted.")
	return nil
}

// ------------------------------------------------------------------------------------------------------------------------------configmap
func createconfigmap(Namespace string) error { //create configmap
	fmt.Println("[0/3]try to create new configmap...")

	configmapClient := clientset.CoreV1().ConfigMaps(Namespace) //create configmap client for api
	fmt.Println("[1/3]client set acquired")

	configmap := &apiv1.ConfigMap{ //declare configmap
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConfigMapName,
			Namespace: Namespace,
		},
		Data: map[string]string{
			"share": "/exports *(rw,fsid=0,insecure,no_root_squash)",
		},
	}

	fmt.Println("[2/3]configmap declared")
	result, err := configmapClient.Create(context.TODO(), configmap, metav1.CreateOptions{}) //create configmap
	if err != nil {
		fmt.Println(err)
		return ErrCantCreateConfigMap
	}
	fmt.Printf("[3/3]configmap created: \"%q\".\n", result.GetObjectMeta().GetName())
	return nil
}

func getConfigMapDEV(Namespace string) (*apiv1.ConfigMapList, error) { //get namespaces function

	fmt.Println("Listing cm: ")
	cmClient := clientset.CoreV1().ConfigMaps(Namespace) //create namespace client for api

	list, err := cmClient.List(context.TODO(), metav1.ListOptions{}) //get namespace list
	if err != nil {
		fmt.Println(err)
		return nil, ErrCantGetConfigmap
	}

	return list, nil
}

/*
func deleteconfigmap(Namespace string) error { //delete cofigmap function(namespace target)
	fmt.Println("Deleting configmap...")
	configmapClient := clientset.CoreV1().ConfigMaps(Namespace)                           //create configmap client for k3s api
	deletePolicy := metav1.DeletePropagationForeground                                    //set delete policy
	if err := configmapClient.Delete(context.TODO(), ConfigMapName, metav1.DeleteOptions{ //delete configmap
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
		return ErrCantDeleteConfigMap
	}
	fmt.Println("configmap deleted.")
	return nil
}
*/
// -------------------------------------------------------------------------------------------------------------------------------api rest functions

var explain = []Response{
	{Message: "Welcome in my api, in the following messages you can see different option"},
	{Message: "WARNING all instruction behind ** chars must be replaced with your data"},
	{Message: "For get your storage servers send a GET request to: /apik3s/storage/*namespace*"},
	{Message: "For get all workspace send a GET request to: /apik3s/storage/"},
	{Message: "For create a new nfs server send a POST request to: /apik3s/storage/*namespace*"},
	{Message: `WARNING using POST request for create a new nfs server require an json file with the following structure:`},
	{Message: `{"id":"*your nfs server id*","size":"*your nfs storage size in GB*","ip":"*server ip*"}`},
	{Message: `EX: {"id": "testserver","size": 3,"ip": "10.2.15.224"}`},
	{Message: `for auto assign ip insert : "auto" or ""`},
	{Message: "For delete an storage server send a DELETE request to: /apik3s/storage/*insert namespace*/*insert server id*"},
	{Message: "For delete all storage server in a workspace send a DELETE request to: /apik3s/storage/*insert namespace*"},
	{Message: "For increment an storage server size send a PATCH request to: /apik3s/storage/*insert namespace*/*insert server id*"},
	{Message: `WARNING using PATCH request for increment an storage server size require an json file with the following structure:`},
	{Message: `{"size":"*your storage size in GB*"}`},
	{Message: `EX: {"size": 5}`},
	{Message: "For get all IPs pool send a GET request to: /apik3s/IPs/"},
	{Message: "For create a IPs pool send a POST request to: /apik3s/IPs/"},
	{Message: `WARNING using POST request for create a IPsPool require an json file with the following structure:`},
	{Message: `{"id":"*your ips pool name*,[*ip address list*]}`},
	{Message: `EX: {"id":"mypool","ips":["192.168.1.12-192.168.1.54","10.2.15.141-10.2.15.143"]}`},
}

type DeleteResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
}

type Response struct {
	Message string `json:"message"`
}

type Workspace struct {
	Name string `json:"name"`
}

type IPpool struct {
	ID  string   `json:"id"`
	Ips []string `json:"ips"`
}

type CreateNFSRequests struct {
	ID   string `json:"id"`
	Size int    `json:"size"`
	IP   string `json:"ip"`
}

type ExtendRequests struct {
	Size int `json:"size"`
}

type NfsStorage struct {
	ID   string `json:"id"`
	Size string `json:"size"`
	IP   string `json:"ip"`
}

/*
simple return every possible operation in a json file
*/
func show(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, explain)
}

/*
this function return a list of every namespace active in k3s cluster
*/
func getWorkspace(context *gin.Context) {
	// get namespaces list
	namespaces, err := getnamespacesDEV()
	if err != nil {
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}
	//declare workspaces array with Workspace type
	var workspaces []Workspace
	for _, namespace := range namespaces.Items { //for every namespace in namespace list if it isn't a default or a system namespace add it in workspaces array
		if namespace.Name != "kube-system" && namespace.Name != "kube-public" && namespace.Name != "kube-node-lease" && namespace.Name != "default" && namespace.Name != "metallb-system" && namespace.Name != "longhorn-system" {
			workspaces = append(workspaces, Workspace{Name: namespace.Name})
		}
	}
	context.IndentedJSON(http.StatusOK, workspaces)
}

/*
this function return every storage actualy active for a namespace
*/
func getStorages(context *gin.Context) {
	//get namespace and service actives

	namespace := context.Param("namespace")
	ServiceList, err := getservicesDEV(namespace)

	if err != nil {
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}
	//get all pvc in a namespace
	PvcList, err := getpvcs(namespace)

	if err != nil {
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}
	//declare servicesJSON var with NfsStorage array type for save all service in json format
	var servicesJSON []NfsStorage

	for _, service := range ServiceList.Items { //for every service in services list
		if service.Name != "kubernetes" { //if service name is different by kubernetes
			var newStorage NfsStorage
			newStorage.ID = service.Name        //create a newStorage With NfsStorage type and set his ID
			for _, pvc := range PvcList.Items { //for every pvc in pvc list
				if pvc.Name == service.Name { //if pvc name match service name
					newStorage.Size = pvc.Spec.Resources.Requests.Storage().String() //set newStorage size by pvc
				}
			}
			for i := 0; i < len(service.Status.LoadBalancer.Ingress); i++ { //for every service ip
				newStorage.IP = service.Status.LoadBalancer.Ingress[i].IP //set ip in newStorage ip
			}
			servicesJSON = append(servicesJSON, newStorage) //append this storage in servicesJSON array
		}
	}
	context.IndentedJSON(http.StatusOK, servicesJSON) //finnaly return the services array
}

/*
this function is used for create a new nfs storage
*/
func addNFSStorage(context *gin.Context) {
	givenNamespace := context.Param("namespace") //get storage namespace by url

	var request CreateNFSRequests //create a new storage request

	if err := context.BindJSON(&request); err != nil { //bind storage request by json input
		context.IndentedJSON(http.StatusBadRequest, Response{Message: ErrBadJsonFormat.Error()})
		return
	}

	if request.Size < 1 { //check requested size
		context.IndentedJSON(http.StatusBadRequest, Response{Message: ErrInvalidSize.Error()})
		return
	}

	size := strconv.Itoa(request.Size) + "Gi"

	namespaceList, err := getnamespacesDEV() //get all namespace
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, Response{Message: err.Error()})
		return
	}

	configMapList, err := getConfigMapDEV(givenNamespace) //get all configmap in given namespace
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, Response{Message: err.Error()})
		return
	}

	var configMapCheck bool = false //declare a bool var used for check in a configmap in this namespace is already created

	for _, configMap := range configMapList.Items { //for all confimap in configmap list
		if configMap.Name == ConfigMapName { //if configmap already exist
			configMapCheck = true //set configmap check variable true
		}
	}

	for _, namespace := range namespaceList.Items { //for all namespace in namespace list
		if namespace.Name == givenNamespace { //if namespace is equal to the given namespace
			if !configMapCheck { //if configmap check is false
				if err := createconfigmap(givenNamespace); err != nil { //create a new confimap
					context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
					return
				}
			} //if configmap check is true
			if err := createpvc(givenNamespace, request.ID, size); err != nil { //create a pvc
				context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
				return
			}
			if err := createNFSdeploy(givenNamespace, request.ID, request.ID); err != nil { //create an nfs deploy
				context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
				deletepvc(givenNamespace, request.ID) //if cant create deploy remove pvc
				return
			}
			if err := createNFSservice(request.ID, givenNamespace, request.ID, request.IP); err != nil { //create an nfs service
				context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
				deletedeploy(givenNamespace, request.ID) //if cant create service remove deploy
				deletepvc(givenNamespace, request.ID)    //if cant create service remove pvc
				return
			}
			context.IndentedJSON(http.StatusCreated, request) //return the request
			return
		}
	} //if the given namespace is not found

	if err := createnamespace(givenNamespace); err != nil { //create a new namespace
		context.IndentedJSON(http.StatusInternalServerError, ErrCantCreateNamespace)
		return
	}

	if err := createconfigmap(givenNamespace); err != nil { //create a new configmap
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		if err := deletenamespace(givenNamespace); err != nil { //if cant create confimap remove namespace
			context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		}
		return
	}

	if err := createpvc(givenNamespace, request.ID, size); err != nil { //create a new pvc
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		if err := deletenamespace(givenNamespace); err != nil { //if cant create pvc remove namespace
			context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		}
		return
	}
	if err := createNFSdeploy(givenNamespace, request.ID, request.ID); err != nil { //create a new nfs deploy
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		if err := deletenamespace(givenNamespace); err != nil { //if cant create deploy remove namespace
			context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		}
		return
	}
	if err := createNFSservice(request.ID, givenNamespace, request.ID, request.IP); err != nil { //create a new nfs service
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		if err := deletenamespace(givenNamespace); err != nil { //if cant create service remove namespace
			context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		}
		return
	}
	context.IndentedJSON(http.StatusCreated, request) //return request
}

/*
this function is used for delete an storage
*/
func deleteStorage(context *gin.Context) {
	namespace := context.Param("namespace") // get storage namespace by url
	id := context.Param("id")               // get storage id by url

	if err := deleteservice(namespace, id); err != nil { //delete storage service
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}
	if err := deletedeploy(namespace, id); err != nil { //delete storage deploy
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}
	if err := deletepvc(namespace, id); err != nil { //delete storage pvc
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}

	context.IndentedJSON(http.StatusOK, DeleteResponse{ID: id, Message: "storage deleted"})
}

/*
this function is used for remove a namespace and all his ccontent
*/
func deleteWorkspace(context *gin.Context) {
	namespace := context.Param("namespace") // get namspace by url

	if err := deletenamespace(namespace); err != nil { //delete namespace
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}

	context.IndentedJSON(http.StatusOK, DeleteResponse{ID: namespace, Message: "namespace deleted"})
}

/*
this function is used to extend an existing storage
*/
func extendStorage(context *gin.Context) {
	namespace := context.Param("namespace") // get storage namespace by url
	id := context.Param("id")               // get storage id by url

	pvc, err := getpvcsbyid(namespace, id) // get specific pvc by id
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, Response{Message: err.Error()})
		return
	}

	//declare an request variable with ExtendRequest type
	var request ExtendRequests

	if err := context.BindJSON(&request); err != nil { // bind request variable with input json
		context.IndentedJSON(http.StatusBadRequest, Response{Message: ErrBadJsonFormat.Error()})
		return
	}

	pvcsize := pvc.Spec.Resources.Requests.Storage().String() //get in string format current pvc size
	pvcsize = strings.TrimRight(pvcsize, "Gi")                //remove Gi from string
	intpvcsize, _ := strconv.Atoi(pvcsize)                    //convert pvc size in int

	if request.Size <= intpvcsize { //if requested pvc size is smallest or equal to old pvc size
		context.IndentedJSON(http.StatusBadRequest, Response{Message: ErrSizeToSmall.Error()}) //return an error
		return
	}

	if err := updatepvc(namespace, id, strconv.Itoa(request.Size)+"Gi"); err != nil { //update pvc size
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}

	context.IndentedJSON(http.StatusOK, Response{Message: "updated to " + strconv.Itoa(request.Size) + "GB"}) //return an success message
}

/*
this function return all current ip pool applied to metallb configuration
*/
func getIPs(context *gin.Context) {
	ipsList, err := getIPAddressPoolsDEV() //get all ip pool list
	if err != nil {
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}
	//declare an pools array variable with IPpool type
	var pools []IPpool
	for _, ips := range ipsList.Items { //for every ip pool in ip pool list
		pools = append(pools, IPpool{ID: ips.Name, Ips: ips.Spec.Addresses}) //append this ip pool to the ip pool array
	}
	context.IndentedJSON(http.StatusOK, pools) //return the array such a json
}

/*
this function is used for create a new ip pool
*/
func addIPs(context *gin.Context) {

	//declare a new ip pool variable with IPpool type
	var requestedPool IPpool

	//bind new ip pool with input json
	if err := context.BindJSON(&requestedPool); err != nil {
		context.IndentedJSON(http.StatusBadRequest, Response{Message: ErrBadJsonFormat.Error()})
		return
	}

	//get current ip pool list
	ipsList, err := getIPAddressPoolsDEV()
	if err != nil {
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}

	for _, ips := range ipsList.Items { //for every ip pool in ip pool list
		if ips.Name == requestedPool.ID { // if ip pool is equal to requested ip pool
			context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()}) //return an error message
			return
		}
	} //else

	if err := createIPAdressPool(requestedPool.ID, requestedPool.Ips); err != nil { //create a new ip pool
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}

	if err := createL2Advertisement(requestedPool.ID); err != nil { //create a new layer 2 advertisment
		if err := deleteIPAddressPool(requestedPool.ID); err != nil { //if cant create layer 2 advertisment delete ip pool
			context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		}
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}
	context.IndentedJSON(http.StatusCreated, requestedPool)
}

/*
this function is used for delete an ip pool
*/
func deleteIPs(context *gin.Context) {
	ID := context.Param("id") //get ip pool name by url

	if err := deleteIPAddressPool(ID); err != nil { //delete ip pool
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}

	if err := deleteL2Advertisement(ID); err != nil { //delete layer 2 advertisment
		context.IndentedJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}
	context.IndentedJSON(http.StatusOK, Response{Message: "ips deleted"})
}

// -------------------------------------------------------------------------------------------------------------------------------main function
func main() {
	//import k3s configuration file locate on local host if it doesn't exist panic
	config, err := clientcmd.BuildConfigFromFlags("", ConfigFilePath)
	if err != nil {
		panic(err.Error())
	}
	//create a metallb client by imported configuration file, if cant create client panic
	metallbClientSet, err = v1beta1.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	//create a kubernetes client by imported configuration file, if cant create client panic
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	_ = clientset
	/*
		create gin engine and se every path function,
		next start gin engine listen.
	*/
	router := gin.Default()
	router.GET("/", show)
	router.GET("/apik3s/storage/:namespace", getStorages)
	router.GET("/apik3s/storage/", getWorkspace)
	router.GET("/apik3s/IPs/", getIPs)
	router.POST("/apik3s/storage/:namespace", addNFSStorage)
	router.POST("/apik3s/IPs/", addIPs)
	router.DELETE("/apik3s/IPs/:id", deleteIPs)
	router.DELETE("/apik3s/storage/:namespace/:id", deleteStorage)
	router.DELETE("/apik3s/storage/:namespace", deleteWorkspace)
	router.PATCH("/apik3s/storage/:namespace/:id", extendStorage)
	router.Run("0.0.0.0:8888")
}
