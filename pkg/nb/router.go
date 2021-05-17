package nb

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// APIRouter should be able to map noobaa api names to actual addresses
// See implementations below: APIRouterNodePort, APIRouterPodPort, APIRouterServicePort
type APIRouter interface {
	GetAddress(api string) string
}

// APIRouterPortForward uses portforwarding to the the pod
type APIRouterPortForward struct {
	ServiceMgmt  *corev1.Service
	PodNamespace string
	PodName      string
	// Start() will setup these fields:
	PF                   *portforward.PortForwarder
	StopChan             chan struct{}
	MapRemotePortToLocal map[uint16]uint16
}

// APIRouterPodPort uses the service target port to route to PodIP:TargetPort
type APIRouterPodPort struct {
	ServiceMgmt *corev1.Service
	PodIP       string
}

// APIRouterNodePort uses the service node port to route to NodeIP:NodePorts
type APIRouterNodePort struct {
	ServiceMgmt *corev1.Service
	NodeIP      string
}

// APIRouterServicePort uses the service port to route to Srv.Namespace:Port
type APIRouterServicePort struct {
	ServiceMgmt *corev1.Service
}

// SimpleRouter is a basic router
type SimpleRouter struct {
	Address string
}

// GetAddress implements the router
func (r *SimpleRouter) GetAddress(api string) string {
	return r.Address
}

var _ APIRouter = &SimpleRouter{}

var forwardingLogRE = regexp.MustCompile(`^Forwarding from 127.0.0.1:(\d+) -> (\d+)$`)

// Start initializes and runs portforwarding by listening on to local ports
// and forwarding their connections to the target pod ports
// See
func (r *APIRouterPortForward) Start() error {

	var ports []string
	for _, p := range r.ServiceMgmt.Spec.Ports {
		ports = append(ports, fmt.Sprintf("0:%s", p.TargetPort.String()))
	}

	config := *util.KubeConfig()
	config.GroupVersion = &schema.GroupVersion{Group: "api", Version: "v1"}
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	restClient, err := rest.RESTClientFor(&config)
	util.Panic(err)

	req := restClient.Post().
		Resource("pods").
		Namespace(r.PodNamespace).
		Name(r.PodName).
		SubResource("portforward")

	var pf *portforward.PortForwarder
	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{}, 1)
	errChan := make(chan error, 1)

	// defer a cleanup handler
	defer func() {
		if stopChan != nil && r.StopChan == nil {
			close(stopChan)
		}
		if pf != nil && r.PF == nil {
			pf.Close()
		}
	}()

	// create spdy dialer (like kubectl)
	transport, upgrader, err := spdy.RoundTripperFor(&config)
	util.Panic(err)
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())

	// create portforwarder
	outLogs := new(bytes.Buffer)
	errLogs := new(bytes.Buffer)
	pf, err = portforward.New(dialer, ports, stopChan, readyChan, outLogs, errLogs)
	util.Panic(err)

	go func() { errChan <- pf.ForwardPorts() }()

	select {
	case <-readyChan:
	case err = <-errChan:
		util.Panic(err)
	}

	// GetPorts doesn't return the actual listener port due to a bug that was fixed in newer version of client-go
	// So instead we have to parse that from the forwarder log output

	// actualPorts, err := pf.GetPorts()
	// util.Panic(err)
	// portMap := make(map[uint16]uint16)
	// for _, p := range actualPorts {
	// 	portMap[p.Remote] = p.Local
	// }

	portMap := make(map[uint16]uint16)
	lines := strings.Split(outLogs.String(), "\n")
	for _, line := range lines {
		matches := forwardingLogRE.FindStringSubmatch(line)
		if matches != nil {
			localPort := matches[1]
			remotePort := matches[2]
			localPortUInt, err := strconv.ParseUint(localPort, 10, 16)
			util.Panic(err)
			remotePortUInt, err := strconv.ParseUint(remotePort, 10, 16)
			util.Panic(err)
			portMap[uint16(remotePortUInt)] = uint16(localPortUInt)
			// fmt.Printf("portMap: %d => %d\n", localPortUInt, remotePortUInt)
		}
	}

	r.PF = pf
	r.StopChan = stopChan
	r.MapRemotePortToLocal = portMap

	return nil
}

// Stop the port forwarding
func (r *APIRouterPortForward) Stop() {
	close(r.StopChan)
	r.PF.Close()
}

// GetAddress implements the router
func (r *APIRouterPortForward) GetAddress(api string) string {
	port := FindPortByName(r.ServiceMgmt, GetAPIPortName(api)).TargetPort.IntValue()
	localPort := r.MapRemotePortToLocal[uint16(port)]
	return fmt.Sprintf("wss://localhost:%d/rpc/", localPort)
}

// GetAddress implements the router
func (r *APIRouterPodPort) GetAddress(api string) string {
	port := FindPortByName(r.ServiceMgmt, GetAPIPortName(api)).TargetPort.IntValue()
	return fmt.Sprintf("wss://%s:%d/rpc/", r.PodIP, port)
}

// GetAddress implements the router
func (r *APIRouterNodePort) GetAddress(api string) string {
	port := FindPortByName(r.ServiceMgmt, GetAPIPortName(api)).NodePort
	return fmt.Sprintf("wss://%s:%d/rpc/", r.NodeIP, port)
}

// GetAddress implements the router
func (r *APIRouterServicePort) GetAddress(api string) string {
	port := FindPortByName(r.ServiceMgmt, GetAPIPortName(api)).Port
	return fmt.Sprintf("wss://%s.%s.svc.cluster.local:%d/rpc/", r.ServiceMgmt.Name, r.ServiceMgmt.Namespace, port)
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
