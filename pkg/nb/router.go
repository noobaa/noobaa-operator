package nb

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// APIRouter should be able to map noobaa api names to actual addresses
// See implementations below: APIRouterNodePort, APIRouterPodPort, APIRouterServicePort
type APIRouter interface {
	GetAddress(api string) string
}

// APIRouterNodePort uses the service node port to route to NodeIP:NodePorts
type APIRouterNodePort struct {
	ServiceMgmt *corev1.Service
	NodeIP      string
}

// APIRouterPodPort uses the service target port to route to PodIP:TargetPort
type APIRouterPodPort struct {
	ServiceMgmt *corev1.Service
	PodIP       string
}

// APIRouterServicePort uses the service port to route to Srv.Namespace:Port
type APIRouterServicePort struct {
	ServiceMgmt *corev1.Service
}

// GetAddress implements the router
func (r *APIRouterNodePort) GetAddress(api string) string {
	port := FindPortByName(r.ServiceMgmt, GetAPIPortName(api)).NodePort
	return fmt.Sprintf("https://%s:%d/rpc/", r.NodeIP, port)
}

// GetAddress implements the router
func (r *APIRouterPodPort) GetAddress(api string) string {
	port := FindPortByName(r.ServiceMgmt, GetAPIPortName(api)).TargetPort.IntValue()
	return fmt.Sprintf("https://%s:%d/rpc/", r.PodIP, port)
}

// GetAddress implements the router
func (r *APIRouterServicePort) GetAddress(api string) string {
	port := FindPortByName(r.ServiceMgmt, GetAPIPortName(api)).Port
	return fmt.Sprintf("https://%s.%s:%d/rpc/", r.ServiceMgmt.Name, r.ServiceMgmt.Namespace, port)
}

// FindPortByName returns the port in the service that matches the given name.
func FindPortByName(srv *corev1.Service, portName string) *corev1.ServicePort {
	for _, p := range srv.Spec.Ports {
		if p.Name == portName {
			return &p
		}
	}
	return &corev1.ServicePort{}
}

// GetAPIPortName maps every noobaa api name to the service port name that serves it.
func GetAPIPortName(api string) string {
	if api == "object_api" || api == "func_api" {
		return "md-https"
	}
	if api == "scrubber_api" {
		return "bg-https"
	}
	if api == "hosted_agents_api" {
		return "hosted-agents-https"
	}
	return "mgmt-https"
}
