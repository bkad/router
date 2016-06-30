package main

import (
	"log"
	"reflect"

	"github.com/drud/router/caddy"
	"github.com/drud/router/model"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util"
)

func main() {
	err := caddy.Start()
	if err != nil {
		log.Fatalf("Failed to start caddy: %v", err)
	}
	kubeClient, err := client.NewInCluster()
	if err != nil {
		log.Fatalf("Failed to create client: %v.", err)
	}
	rateLimiter := util.NewTokenBucketRateLimiter(0.1, 1)
	known := &model.RouterConfig{}
	// Main loop
	for {
		rateLimiter.Accept()
		routerConfig, err := model.Build(kubeClient)
		if err != nil {
			log.Printf("Error building model; not modifying certs or configuration: %v.", err)
			continue
		}
		if reflect.DeepEqual(routerConfig, known) {
			continue
		}
		log.Println("INFO: Router configuration has changed in k8s.")
		err = caddy.WriteCerts(routerConfig, "/opt/router/ssl")
		if err != nil {
			log.Printf("Failed to write certs; continuing with existing certs, dhparam, and configuration: %v", err)
			continue
		}
		err = caddy.WriteConfig(routerConfig, "/opt/router/Caddyfile")
		if err != nil {
			log.Printf("Failed to write new caddy configuration; continuing with existing configuration: %v", err)
			continue
		}
		err = caddy.Reload()
		if err != nil {
			log.Fatalf("Failed to reload caddy: %v", err)
		}
		known = routerConfig
	}
}
