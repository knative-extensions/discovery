package main

import (
	"log"

	"knative.dev/hack/schema/commands"
	"knative.dev/hack/schema/registry"

	"knative.dev/discovery/pkg/apis/discovery/v1alpha1"
)

// This is a demo of what the CLI looks like, copy and implement your own.
func main() {
	registry.Register(&v1alpha1.ClusterDuckType{})

	if err := commands.New("knative.dev/discovery").Execute(); err != nil {
		log.Fatal("Error during command execution: ", err)
	}
}
