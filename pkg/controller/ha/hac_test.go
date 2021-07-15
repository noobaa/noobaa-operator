package hac_test

import (
	"context"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

var _ = Describe("High Availability (HA) integration test", func() {
	// Expect a 5nodes KIND cluster,
	// see .travis/install-5nodes-kind-cluster.sh
	const (
		nodesNum = 5
	)

	// HAC test variables
	var (
		ctx            = context.Background() // TODO context ;)
		nodes          *v1.NodeList           // Initial cluster nodes list
		pods           *v1.PodList            // NooBaa pods running in the cluster
		nodeToKill     *string                // Node selected to be killed
		podsToEvictMap = map[string]bool{}    // Set of NooBaa pods expected
		                                      // to be evicted as a result of test
		err            error                  // Test err
	)

	// - Verify the integratation test cluster environment
	// - Choose a cluster node to stop
	// - Calculate the set of NooBaa pods to be evicted
	Context("Verify K8S cluster", func() {

		// Verify 5 nodes cluster
		Specify("Require 5 node cluster", func() {
			nodes, err = clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(nodes).ToNot(BeNil())

			for _, n := range nodes.Items {
				logger.Printf("found node %q", n.Name)
			}

			Expect(len(nodes.Items)).To(BeIdenticalTo(nodesNum), "kind 5 noodes cluster is expected")
		})

		// Verify NooBaa installation
		Specify("Require NooBaa pods", func() {
			labelOption := metav1.ListOptions{LabelSelector: "app=noobaa"}
			pods, err = clientset.CoreV1().Pods(metav1.NamespaceDefault).List(ctx, labelOption)
			Expect(err).ToNot(HaveOccurred())
			Expect(pods).ToNot(BeNil())

			for _, p := range pods.Items {
				logger.Printf("found NooBaa pod %v on node %v", p.Name, p.Spec.NodeName)
			}

			Expect(len(pods.Items)).ToNot(BeIdenticalTo(0))
		})

		// Select a node to stop and
		// calculate list of NooBaa pods expected to be evicted
		Specify("Require a node to kill", func() {
			nodeByPodPrefix := func(pref string) string {
				for _, p := range pods.Items {
					if strings.HasPrefix(p.Name, pref) {
						return p.Spec.NodeName
					}
				}
				return ""
			}

			// Select a node to stop such as:
			// - populated by NooBaa pod
			// - operator not running on this node
			podsToEvictPrefix := []string{
				"noobaa-core",
				"noobaa-endpoint",
				"noobaa-" + metav1.NamespaceDefault + "-backing-store",
			}
			operatorPref := "noobaa-operator"
			for _, pref := range podsToEvictPrefix {
				candidateNode := nodeByPodPrefix(pref)
				if len(candidateNode) > 0 && nodeByPodPrefix(operatorPref) != candidateNode {
					nodeToKill = &candidateNode
				}
			}
			Expect(nodeToKill).ToNot(BeNil())
			Expect(len(*nodeToKill) > 0).To(BeTrue())

			logger.Printf("node to kill %q", *nodeToKill)

			// Calculate a set of NooBaa pods
			// expected to be deleted by the operator
			listOption := metav1.ListOptions{LabelSelector: "app=noobaa", FieldSelector: "spec.nodeName=" + (*nodeToKill)}
			podsToEvict, err := clientset.CoreV1().Pods(metav1.NamespaceDefault).List(ctx, listOption)
			Expect(err).ToNot(HaveOccurred())
			Expect(podsToEvict).ToNot(BeNil())

			logger.Printf("Pods to be evicted")
			for _, p := range podsToEvict.Items {
				logger.Printf("   %q", p.Name)
				podsToEvictMap[p.Name] = true
			}

			Expect(len(podsToEvictMap)).NotTo(BeIdenticalTo(0))
		})
	})

	// Node failure flow:
	// - Initiate a node failue by stopping worker node container
	// - Wait for NooBaa pods eviction from the failing node
	Context("Node Failure", func() {

		// Initiate node failure
		Specify("Require docker stop success", func() {
			Expect(nodeToKill).ToNot(BeNil())
			Expect(len(*nodeToKill) > 0).To(BeTrue())

			cmd := exec.Command("docker", "stop", *nodeToKill)
			err = cmd.Run()
			Expect(err).ToNot(HaveOccurred())
		})

		// Verify NooBaa pods were evicted
		Specify("Require all pods to be deleted", func() {
			Expect(nodeToKill).ToNot(BeNil())
			Expect(len(*nodeToKill) > 0).To(BeTrue())

			listOption := metav1.ListOptions{LabelSelector: "app=noobaa", FieldSelector: "spec.nodeName=" + (*nodeToKill)}
			w, err := clientset.CoreV1().Pods(metav1.NamespaceDefault).Watch(ctx, listOption)
			Expect(err).ToNot(HaveOccurred())

			timeoutDuration := 5 * time.Minute
			timeoutTime := time.Now().Add(timeoutDuration)

			logger.Printf("Waiting for pods to be evicted %v", podsToEvictMap)
			for len(podsToEvictMap) > 0 && timeoutTime.After(time.Now()) {
				select {
				case e := <-w.ResultChan():
					if e.Type != watch.Deleted {
						continue
					}
					pod := e.Object.(*v1.Pod)
					delete(podsToEvictMap, pod.Name)
					logger.Printf("evicted  %q", pod.Name)
				case <-time.After(time.Until(timeoutTime)):
				}
			}

			Expect(len(podsToEvictMap)).To(BeIdenticalTo(0))
		})
	})
})
