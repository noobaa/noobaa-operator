package bench

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EndpointType = string

const (
	EndpointInternal     EndpointType = "internal"
	EndpointPodIP        EndpointType = "podip"
	EndpointNodeport     EndpointType = "nodeport"
	EndpointLoadbalancer EndpointType = "loadbalancer"
	EndpointManual       EndpointType = "manual"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bench",
		Short: "Run benchmark",
	}
	cmd.AddCommand(
		CmdWarp(),
	)

	return cmd
}

func CmdWarp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "warp",
		Short: "Run warp benchmark",
		Run:   RunBenchWarp,
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().String(
		"bucket", "first.bucket",
		"Bucket to use for benchmark data. ALL DATA WILL BE DELETED IN BUCKET!",
	)
	cmd.Flags().Bool(
		"use-https", false,
		"Use HTTPS endpoints for benchmark",
	)
	cmd.Flags().Int32(
		"clients", 0,
		"Number of warp instances",
	)
	cmd.Flags().String(
		"image", "minio/warp:latest",
		"Warp image",
	)
	cmd.Flags().String(
		"endpoint-type", EndpointInternal,
		fmt.Sprintf(
			"Endpoint type could be %s,%s,%s,%s,%s",
			EndpointInternal, EndpointPodIP, EndpointNodeport,
			EndpointLoadbalancer, EndpointManual,
		),
	)
	cmd.Flags().String(
		"access-key", "",
		"Access Key to access the S3 bucket",
	)
	cmd.Flags().String(
		"secret-key", "",
		"Secret Key to access the S3 bucket",
	)
	cmd.Flags().String(
		"warp-args", "",
		"Arguments to be passed directly to warp CLI",
	)

	return cmd
}

func RunBenchWarp(cmd *cobra.Command, args []string) {
	log := util.Logger()
	log.Info("Starting warp benchmark")

	bench := args[0]
	bucket := util.GetFlagStringOrPrompt(cmd, "bucket")
	image := util.GetFlagStringOrPrompt(cmd, "image")
	endpointType := util.GetFlagStringOrPrompt(cmd, "endpoint-type")

	useHTTPs, _ := cmd.Flags().GetBool("use-https")
	clients, _ := cmd.Flags().GetInt32("clients")
	accessKey, _ := cmd.Flags().GetString("access-key")
	secretKey, _ := cmd.Flags().GetString("secret-key")
	warpArgs, _ := cmd.Flags().GetString("warp-args")

	if accessKey == "" || secretKey == "" {
		noobaaAdminSecret := util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)
		noobaaAdminSecret.Name = "noobaa-admin"
		noobaaAdminSecret.Namespace = options.Namespace

		if !util.KubeCheck(noobaaAdminSecret) {
			log.Fatal("❌ Access Key and/or Secret Key not provided and failed to fetch \"noobaa-admin\" secret")
		}

		accessKey = noobaaAdminSecret.StringData["AWS_ACCESS_KEY_ID"]
		secretKey = noobaaAdminSecret.StringData["AWS_SECRET_ACCESS_KEY"]

		if accessKey == "" || secretKey == "" {
			log.Fatal("❌ Access Key and/or Secret Key not provided and failed to find credentials in \"noobaa-admin\" secret")
		}
	}

	clients = getClientNums(clients)
	if clients == 0 {
		log.Fatal("❌ Number of clients cannot be 0")
	}

	// Validate "bench" type
	allowedBench := []string{
		"mixed", "get", "put",
		"delete", "list", "stat",
		"versioned", "multipart",
		"multipart-put",
	}
	if !slices.Contains(allowedBench, bench) {
		log.Fatalf("❌ Invalid bench provided - supported %+v", allowedBench)
	}

	// Validate endpoint type
	allowedEndpointType := []EndpointType{
		EndpointInternal, EndpointLoadbalancer, EndpointNodeport,
		EndpointPodIP, EndpointManual,
	}
	if !slices.Contains(allowedEndpointType, endpointType) {
		log.Fatalf("❌ Invalid endpoint-type provided - supported %+v", allowedEndpointType)
	}

	go util.OnSignal(func() {
		cleanupWarp()
	}, syscall.SIGINT, syscall.SIGTERM)

	provisionWarp(clients, image)
	startWarpJob(
		clients,
		bench, bucket, accessKey, secretKey, image,
		endpointType, useHTTPs, warpArgs,
	)
	pollWarp()
	cleanupWarp()
}

func provisionWarp(clients int32, image string) {
	log := util.Logger()

	warpSts := util.KubeObject(bundle.File_deploy_warp_warp_yaml).(*appsv1.StatefulSet)
	containers := warpSts.Spec.Template.Spec.Containers
	if len(containers) != 1 {
		log.Fatal("❌ Unexepected number of containers in the Warp STS")
	}

	// Provision in the same namespace where NooBaa is running
	warpSts.Namespace = options.Namespace
	warpSts.Spec.Replicas = &clients
	if image != "" {
		containers[0].Image = image
	}
	util.KubeApply(warpSts)

	// Wait for the warp client to get ready
	if err := wait.PollUntilContextCancel(context.TODO(), 2*time.Second, true, func(ctx context.Context) (done bool, err error) {
		util.KubeCheck(warpSts)
		return warpSts.Status.ReadyReplicas == *warpSts.Spec.Replicas, nil
	}); err != nil {
		log.Fatalf("❌ error while waiting for warp sts to be ready: %s", err)
	}

	warpSvc := util.KubeObject(bundle.File_deploy_warp_warp_svc_yaml).(*corev1.Service)
	warpSvc.Namespace = options.Namespace
	util.KubeApply(warpSvc)
}

func startWarpJob(
	clients int32,
	bench, bucket, accessKey, secretKey, image string,
	endpointType EndpointType,
	useHttps bool,
	warpArgs string,
) {
	warpJob := util.KubeObject(bundle.File_deploy_warp_warp_job_yaml).(*batchv1.Job)
	warpJob.Namespace = options.Namespace

	for idx, container := range warpJob.Spec.Template.Spec.Containers {
		if container.Name == "warp-job" {
			warpJob.Spec.Template.Spec.Containers[idx].Env = []corev1.EnvVar{
				{
					Name:  "WARP_ACCESS_KEY",
					Value: accessKey,
				},
				{
					Name:  "WARP_SECRET_KEY",
					Value: secretKey,
				},
			}

			args := append(
				container.Args,
				bench,
				"--bucket", bucket,
			)
			if endpointType != EndpointManual {
				args = append(
					args,
					"--host", prepareWarpHostList(endpointType, useHttps),
				)
			}
			if clients == 1 {
				args = append(
					args,
					"--warp-client", fmt.Sprintf("warp-0.warp.%s.svc.cluster.local:7761", options.Namespace),
				)
			} else {
				args = append(
					args,
					"--warp-client", fmt.Sprintf("warp-{0..%d}.warp.%s.svc.cluster.local:7761", clients-1, options.Namespace),
				)
			}

			if useHttps {
				args = append(args, "--insecure", "--tls")
			}
			if warpArgs != "" {
				args = append(args, strings.Split(warpArgs, " ")...)
			}

			warpJob.Spec.Template.Spec.Containers[idx].Args = args
			warpJob.Spec.Template.Spec.Containers[idx].Image = image
			break
		}
	}

	util.KubeApply(warpJob)
}

func pollWarp() {
	log := util.Logger()

	log.Infof(
		"Benchmark started - you can check the warp client logs by \"kubectl logs -f -n %s -l app=warp\"\n",
		options.Namespace,
	)

	warpJob := util.KubeObject(bundle.File_deploy_warp_warp_job_yaml).(*batchv1.Job)
	warpJob.Namespace = options.Namespace
	for {
		if !util.KubeCheckQuiet(warpJob) {
			log.Error("❌ No Warp Job found to poll")
			break
		}

		if warpJob.Status.Succeeded == 0 && warpJob.Status.Failed == 0 {
			log.Info("Benchmark still running")
		} else {
			break
		}

		time.Sleep(5 * time.Second)
	}

	if warpJob.Status.Failed != 0 {
		log.Error("❌ Warp Job Failed")
	}

	warpJobPodList := &corev1.PodList{}
	util.KubeList(warpJobPodList, client.InNamespace(options.Namespace), client.MatchingLabels{
		"job-name": "warp-job",
	})

	for _, pod := range warpJobPodList.Items {
		logs, err := util.GetPodLogs(pod)
		if err != nil {
			log.Errorf("❌ Failed to get logs for pod %q - %s", pod.Name, err)
			continue
		}

		for _, container := range logs {
			if _, err := io.Copy(os.Stdout, container); err != nil {
				log.Warn("encountered error while copying logs -", err)
			}
		}
	}
}

func cleanupWarp() {
	util.Logger().Info("Cleaning up Warp")

	warpJob := util.KubeObject(bundle.File_deploy_warp_warp_job_yaml).(*batchv1.Job)
	warpJob.Namespace = options.Namespace
	deletePolicy := metav1.DeletePropagationForeground
	util.KubeDelete(warpJob, &client.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	warpSvc := util.KubeObject(bundle.File_deploy_warp_warp_svc_yaml).(*corev1.Service)
	warpSvc.Namespace = options.Namespace
	util.KubeDelete(warpSvc)

	warpSts := util.KubeObject(bundle.File_deploy_warp_warp_yaml).(*appsv1.StatefulSet)
	warpSts.Namespace = options.Namespace
	util.KubeDelete(warpSts)
}

func prepareWarpHostList(endpointType EndpointType, https bool) string {
	log := util.Logger()
	ips := []string{}
	port := int32(0)

	portname := "s3"
	if https {
		portname = "s3-https"
	}

	getPort := func(svc *corev1.Service, portname string, fn func(corev1.ServicePort) int32) int32 {
		for _, port := range svc.Spec.Ports {
			if port.Name == portname {
				return fn(port)
			}
		}

		return 0
	}

	s3svc := util.KubeObject(bundle.File_deploy_internal_service_s3_yaml).(*corev1.Service)
	s3svc.Namespace = options.Namespace
	s3svc.Spec = corev1.ServiceSpec{}
	if !util.KubeCheck(s3svc) {
		log.Fatalf("❌ Failed to find S3 service in namespace %s", options.Namespace)
	}

	if endpointType == EndpointInternal {
		ips = append(ips, fmt.Sprintf("s3.%s.svc", options.Namespace))
		port = getPort(s3svc, portname, func(sp corev1.ServicePort) int32 {
			return sp.Port
		})
	}
	if endpointType == EndpointPodIP {
		endpoints := corev1.PodList{}
		if !util.KubeList(&endpoints, client.InNamespace(options.Namespace), client.MatchingLabels{"noobaa-s3": "noobaa"}) {
			log.Fatalf("❌ Failed to find endpoint pods in namespace: %s", options.Namespace)
		}

		for _, ep := range endpoints.Items {
			ips = append(ips, ep.Status.PodIP)
		}
		port = getPort(s3svc, portname, func(sp corev1.ServicePort) int32 {
			return sp.TargetPort.IntVal
		})
	}
	if endpointType == EndpointNodeport {
		nodes := corev1.NodeList{}
		if !util.KubeList(&nodes) || len(nodes.Items) == 0 {
			log.Fatalf("❌ Failed to find node IP - failed to find nodes")
		}

		for _, node := range nodes.Items {
			for _, address := range node.Status.Addresses {
				if address.Type == corev1.NodeExternalIP {
					ips = append(ips, address.Address)
				}
			}

			// Use the first external IP we find on the node
			if len(ips) > 0 {
				break
			}
		}
		// Fallback to interalIP if no external found
		if len(ips) == 0 {
			for _, node := range nodes.Items {
				for _, a := range node.Status.Addresses {
					if a.Type == corev1.NodeInternalIP {
						ips = append(ips, a.Address)
					}
				}

				if len(ips) > 0 {
					break
				}
			}
		}
		port = getPort(s3svc, portname, func(sp corev1.ServicePort) int32 {
			return sp.NodePort
		})
	}
	if endpointType == EndpointLoadbalancer {
		if len(s3svc.Status.LoadBalancer.Ingress) == 0 {
			log.Fatal("❌ Failed to find loadbalancer ingress")
		}

		if s3svc.Status.LoadBalancer.Ingress[0].IP != "" {
			ips = append(ips, s3svc.Status.LoadBalancer.Ingress[0].IP)
		} else if s3svc.Status.LoadBalancer.Ingress[0].Hostname != "" {
			ips = append(ips, s3svc.Status.LoadBalancer.Ingress[0].Hostname)
		} else {
			log.Fatal("❌ Failed to find loadbalancer IP/Hostname")
		}
		port = getPort(s3svc, portname, func(sp corev1.ServicePort) int32 {
			return sp.Port
		})
	}

	return strings.Join(util.Map(ips, func(ip string) string {
		return fmt.Sprintf("%s:%d", ip, port)
	}), ",")
}

func getClientNums(clients int32) int32 {
	log := util.Logger()

	nodelist := corev1.NodeList{}
	util.KubeList(&nodelist)
	nodeCount := len(nodelist.Items)
	if nodeCount == 0 {
		log.Fatal("❌ No nodes found to run warp clients")
	}

	if clients > int32(nodeCount) || clients == 0 {
		clients = int32(nodeCount)
		if clients > int32(nodeCount) {
			log.Warn("Number of clients cannot exceed number of nodes")
		}
	}

	return clients
}
