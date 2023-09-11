package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/pointer"
)

const roleName = "k8s-fix-unhealthy-pod"

func main() {
	var ns string
	flag.StringVar(&ns, "ns", "default", "namespace")
	flag.Parse()
	log.Printf("start to check unhealthy pods for namespace [%s]...\n", ns)

	config := getK8sConfig()

	// Create an rest client not targeting specific API version
	k8sclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	var aiclient appinsights.TelemetryClient
	if os.Getenv("APPINSIGHTS_INSTRUMENTATIONKEY") != "" {
		aiclient = createTelemetryClient()
	}

	events, err := getUnhealthyPodEvents(k8sclient, ns)
	if err != nil {
		log.Fatalln("failed to get events:", err)
	}

	log.Printf("will process unhealthy pod events. count is [%d]...\n", len(events.Items))

	// kill unhealthy pods
	for i, event := range events.Items {
		if shouldProcessUnhealthyPod(event) {
			processUnhealthyPod(i, event, k8sclient, ns, aiclient)
		}
	}

	fmt.Println("all proccessed!")
	fmt.Println("bye!")
}

func shouldProcessUnhealthyPod(event corev1.Event) bool {
	return strings.Contains(event.Message, "context deadline exceeded (Client.Timeout exceeded while awaiting headers)")
}

func processUnhealthyPod(i int, event corev1.Event, k8sclient *kubernetes.Clientset, ns string, aiclient appinsights.TelemetryClient) {
	var msg string
	msg = fmt.Sprintf("will process unhealthy pod. name is [%s]...\n", event.InvolvedObject.Name)
	log.Print(msg)
	if aiclient != nil {
		aiclient.TrackTrace(msg, appinsights.Information)
	}
	pod, err := k8sclient.CoreV1().Pods(ns).Get(context.Background(), event.InvolvedObject.Name, metav1.GetOptions{})
	if err != nil {
		log.Println("ERROR: failed to get pod:", err)
		return
	}

	deleteUnhealthyPod(i, pod.Name, k8sclient, ns)
	msg = fmt.Sprintf("processed unhealthy pod. name is [%s]...\n", event.InvolvedObject.Name)
	log.Print(msg)
	if aiclient != nil {
		aiclient.TrackTrace(msg, appinsights.Information)
	}
}

func getUnhealthyPodEvents(k8sclient *kubernetes.Clientset, ns string) (*corev1.EventList, error) {
	eventListOptions := metav1.ListOptions{FieldSelector: "reason=Unhealthy,involvedObject.kind=Pod"}
	return k8sclient.CoreV1().Events(ns).List(context.Background(), eventListOptions)
}

func createTelemetryClient() appinsights.TelemetryClient {
	telemetryConfig := appinsights.NewTelemetryConfiguration(os.Getenv("APPINSIGHTS_INSTRUMENTATIONKEY"))
	if os.Getenv("APPINSIGHTS_ENDPOINTURL") != "" {
		telemetryConfig.EndpointUrl = os.Getenv("APPINSIGHTS_ENDPOINTURL")
	}
	aiclient := appinsights.NewTelemetryClientFromConfig(telemetryConfig)
	aiclient.Context().CommonProperties["AppRoleName"] = roleName
	return aiclient
}

func deleteUnhealthyPod(i int, name string, k8sclient *kubernetes.Clientset, ns string) {
	fmt.Printf("-- [%d] %s pod unhealthy, will be killed\n", i, name)

	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: pointer.Int64(30),
		PropagationPolicy:  &[]metav1.DeletionPropagation{metav1.DeletePropagationBackground}[0],
	}
	err := k8sclient.CoreV1().Pods(ns).Delete(context.Background(), name, deleteOptions)
	if err == nil {
		fmt.Printf("---- [%d] %s killed\n", i, name)
	} else {
		log.Printf("---- ERROR: [%d] %s failed to kill: %s\n", i, name, err)
	}
}

// getK8sConfig returns a kubernetes config from InCluster or config file
func getK8sConfig() *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Fatal(err)
	}
	return config
}
