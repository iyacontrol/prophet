package main

import (
	"flag"
	"os"
	"strconv"

	log "github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	EnvHpaName = "CRONHPA_NAME"
	EnvHpaNamespace = "CRONHPA_NAMESPACE"
	EnvHpaCapacity = "CRONHPA_CAPACITY"
)


func main() {
	flag.Parse()

	name := os.Getenv(EnvHpaName)
	if name == "" {
		log.Fatal("hpa name must supply!")
	}

	namespace := os.Getenv(EnvHpaNamespace)
	if namespace == "" {
		log.Fatal("hpa namespace must supply!")
	}
	capacity := os.Getenv(EnvHpaCapacity)
	if capacity == "" {
		log.Fatal("hpa capacity must supply!")
	}

	cap, err:=strconv.Atoi(capacity)
	if err != nil {
		log.Fatal(err)
	}

	cap32 := int32(cap)


	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}


	hpa, err := clientset.AutoscalingV2beta1().HorizontalPodAutoscalers(namespace).Get(name, v1.GetOptions{})
	if err != nil {
		log.Errorf("can not get hpa: %s in namespace: %s", name, namespace)
		return
	}

	hpa.Spec.MinReplicas = &cap32

	_, err = clientset.AutoscalingV2beta1().HorizontalPodAutoscalers(namespace).Update(hpa)
	if err != nil {
		log.Errorf("can not update hpa: %s in namespace: %s", name, namespace)
		return
	}

	log.Infof("success update hpa: %s in namespace: %s", name, namespace)
}