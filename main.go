package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"reflect"

	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/flowcontrol"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <labelselector> <subnet>\n", os.Args[0])
		os.Exit(0)
	}

	labelselector := os.Args[1]
	subnetinput := os.Args[2]
	_, ipnet, err := net.ParseCIDR(subnetinput)
	if err != nil {
		log.Fatalln(err)
	}
	subnet := ipnet.String()
	config := getKubeConfig()
	clientset := kubernetes.NewForConfigOrDie(config)
	rateLimiter := flowcontrol.NewTokenBucketRateLimiter(0.1, 1)

	log.Println("Listening for Pods with LabelSelector:", labelselector)

	var knowclientIP string
	for {
		rateLimiter.Accept()
		clientIps := getPods(clientset, labelselector)
		if len(clientIps) == 0 {
			log.Println("WARNING: No IPs found")
			continue
		}
		clientIP := clientIps[0]

		if len(clientIps) > 1 {
			log.Println("WARNING: ", len(clientIps), " IPs found")
			for _, ip := range clientIps {
				log.Println("WARNING: IP", ip)
			}
			log.Println("WARNING: Taking IP", clientIP)
		}

		if reflect.DeepEqual(clientIP, knowclientIP) {
			continue
		}

		out, err := exec.Command(
			"ip", "route", "delete", subnet).CombinedOutput()
		// out, err := exec.Command(
		// 	"echo", "ip route delete", subnet).CombinedOutput()
		if err != nil {
			log.Println("ERROR ip route delete:", err)
			continue
		}
		log.Println("IP ROUTE DELETE:", string(out[:]))

		out, err = exec.Command(
			"ip", "route", "add", subnet, "via", clientIP).CombinedOutput()
		// out, err = exec.Command(
		// 	"echo", "ip route add", subnet, "via", clientIP).CombinedOutput()

		if err != nil {
			log.Println("ERROR ip route add:", err)
			continue
		}

		log.Println("IP ROUTE ADD:", string(out[:]))
		log.Println("IP route added for subnet ", subnet, "via", clientIP)
		knowclientIP = clientIP
	}
}

func getKubeConfig() *rest.Config {
	// var kubeconfig *string
	// if home := homeDir(); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }
	// flag.Parse()

	// use the current context in kubeconfig
	// config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	config, err := rest.InClusterConfig()

	if err != nil {
		log.Fatalln(err)
	}
	return config
}

func getPods(clientset *kubernetes.Clientset, labelSelector string) []string {
	var clientIps []string

	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln(err)
	}
	listopt := metav1.ListOptions{LabelSelector: labelSelector}
	for _, ns := range namespaces.Items {
		pods, _ := clientset.CoreV1().Pods(ns.Name).List(listopt)
		for _, pod := range pods.Items {
			clientIps = append(clientIps, pod.Status.PodIP)
		}
	}
	return clientIps
}

// func homeDir() string {
// 	if h := os.Getenv("HOME"); h != "" {
// 		return h
// 	}
// 	return os.Getenv("USERPROFILE") // windows
// }
