package system

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcilePhaseConnecting runs the reconcile phase
func (r *Reconciler) ReconcilePhaseConnecting() error {

	r.SetPhase(
		nbv1.SystemPhaseConnecting,
		"SystemPhaseConnecting",
		"noobaa operator started phase 3/4 - \"Connecting\"",
	)

	if r.JoinSecret == nil {
		r.CheckServiceStatus(r.ServiceMgmt, r.RouteMgmt, &r.NooBaa.Status.Services.ServiceMgmt, "mgmt-https")
	}
	if err := r.InitNBClient(); err != nil {
		return err
	}

	r.CheckServiceStatus(r.ServiceS3, r.RouteS3, &r.NooBaa.Status.Services.ServiceS3, "s3-https")
	r.CheckServiceStatus(r.ServiceSts, r.RouteSts, &r.NooBaa.Status.Services.ServiceSts, "sts-https")

	return nil

	// if len(r.NooBaa.Status.Services.ServiceMgmt.PodPorts) != 0 {
	// 	podPort := r.NooBaa.Status.Services.ServiceMgmt.PodPorts[0]
	// 	podIP := podPort[strings.Index(podPort, "://")+3 : strings.LastIndex(podPort, ":")]
	// 	r.NBClient = nb.NewClient(&nb.APIRouterPodPort{
	// 		ServiceMgmt: r.ServiceMgmt,
	// 		PodIP:       podIP,
	// 	})
	// 	r.NBClient.SetAuthToken(r.SecretOp.StringData["auth_token"])
	// 	return nil
	// }
}

// CheckServiceStatus populates the status of a service by detecting all of its addresses
func (r *Reconciler) CheckServiceStatus(srv *corev1.Service, route *routev1.Route, status *nbv1.ServiceStatus, portName string) {

	log := r.Logger.WithField("func", "CheckServiceStatus").WithField("service", srv.Name)
	*status = nbv1.ServiceStatus{}
	servicePort := nb.FindPortByName(srv, portName)
	proto := "http"
	if strings.HasSuffix(portName, "https") {
		proto = "https"
	}

	// Node IP:Port
	// Pod IP:Port
	pods := corev1.PodList{}
	podsListOptions := &client.ListOptions{
		Namespace:     r.Request.Namespace,
		LabelSelector: labels.SelectorFromSet(srv.Spec.Selector),
	}
	err := r.Client.List(r.Ctx, &pods, podsListOptions)
	if err == nil {
		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodRunning {
				if pod.Status.HostIP != "" {
					status.NodePorts = append(
						status.NodePorts,
						fmt.Sprintf("%s://%s:%d", proto, pod.Status.HostIP, servicePort.NodePort),
					)
				}
				if pod.Status.PodIP != "" {
					status.PodPorts = append(
						status.PodPorts,
						fmt.Sprintf("%s://%s:%s", proto, pod.Status.PodIP, servicePort.TargetPort.String()),
					)
				}
			}
		}
	}

	// Cluster IP:Port (of the service)
	if srv.Spec.ClusterIP != "" {
		status.InternalIP = append(
			status.InternalIP,
			fmt.Sprintf("%s://%s:%d", proto, srv.Spec.ClusterIP, servicePort.Port),
		)
		status.InternalDNS = append(
			status.InternalDNS,
			fmt.Sprintf("%s://%s.%s.svc:%d", proto, srv.Name, srv.Namespace, servicePort.Port),
		)
	}

	// Routes IP:Port (of the service) {
	if route.Spec.Host != "" {
		status.ExternalDNS = append(
			status.ExternalDNS,
			fmt.Sprintf("%s://%s", proto, route.Spec.Host),
		)
	}

	// LoadBalancer IP:Port (of the service)
	if srv.Status.LoadBalancer.Ingress != nil {
		for _, lb := range srv.Status.LoadBalancer.Ingress {
			if lb.IP != "" {
				status.ExternalIP = append(
					status.ExternalIP,
					fmt.Sprintf("%s://%s:%d", proto, lb.IP, servicePort.Port),
				)
			}
			if lb.Hostname != "" {
				status.ExternalDNS = append(
					status.ExternalDNS,
					fmt.Sprintf("%s://%s:%d", proto, lb.Hostname, servicePort.Port),
				)
			}
		}
	}

	// External IP:Port (of the service)
	if srv.Spec.ExternalIPs != nil {
		for _, ip := range srv.Spec.ExternalIPs {
			status.ExternalIP = append(
				status.ExternalIP,
				fmt.Sprintf("%s://%s:%d", proto, ip, servicePort.Port),
			)
		}
	}

	log.Infof("Collected addresses: %+v", status)
}

// InitNBClient initialize the noobaa client for making calls to the server.
func (r *Reconciler) InitNBClient() error {
	if r.JoinSecret == nil {
		r.NBClient = nb.NewClient(&nb.APIRouterServicePort{
			ServiceMgmt: r.ServiceMgmt,
		})

	} else {
		addr := r.JoinSecret.StringData["mgmt_addr"]
		u, err := url.Parse(addr)
		// The URL's Parse method "may not necessarily return an error, due to parsing ambiguities"
		if err != nil {
			return err
		}
		u.Path = path.Join(u.Path, "rpc")
		addr = u.String()
		r.NBClient = nb.NewClient(&nb.SimpleRouter{
			Address: addr,
		})
	}

	// Check that the server is indeed serving the API already
	// we use the read_auth call here because it's an API that always answers
	// even when auth_token is empty.
	_, err := r.NBClient.ReadAuthAPI()
	return err
}
