package cmd

import (
	"bufio"
	"fmt"
	"github.com/cilium/cilium/api/v1/models"
	"github.com/spf13/cobra"
	"net"
	"os"
	"strings"
)

var (
	k8sExternalIPs     bool
	k8sNodePort        bool
	k8sHostPort        bool
	k8sLoadBalancer    bool
	k8sTrafficPolicy   string
	k8sClusterInternal bool
	localRedirect      bool
	idU                uint64
	frontend           string
	backends           []string
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage services & loadbalancers",
}

var serviceUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a service",
	Run: func(cmd *cobra.Command, args []string) {
		updateService(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)

	serviceCmd.AddCommand(serviceUpdateCmd)
	serviceUpdateCmd.Flags().Uint64VarP(&idU, "id", "", 0, "Identifier")
	serviceUpdateCmd.Flags().BoolVarP(&k8sExternalIPs, "k8s-external", "", false, "Set service as a k8s ExternalIPs")
	serviceUpdateCmd.Flags().BoolVarP(&k8sNodePort, "k8s-node-port", "", false, "Set service as a k8s NodePort")
	serviceUpdateCmd.Flags().BoolVarP(&k8sLoadBalancer, "k8s-load-balancer", "", false, "Set service as a k8s LoadBalancer")
	serviceUpdateCmd.Flags().BoolVarP(&k8sHostPort, "k8s-host-port", "", false, "Set service as a k8s HostPort")
	serviceUpdateCmd.Flags().BoolVarP(&localRedirect, "local-redirect", "", false, "Set service as Local Redirect")
	serviceUpdateCmd.Flags().StringVarP(&k8sTrafficPolicy, "k8s-traffic-policy", "", "Cluster", "Set service with k8s externalTrafficPolicy as {Local,Cluster}")
	serviceUpdateCmd.Flags().BoolVarP(&k8sClusterInternal, "k8s-cluster-internal", "", false, "Set service as cluster-internal for externalTrafficPolicy=Local")
	serviceUpdateCmd.Flags().StringVarP(&frontend, "frontend", "", "", "Frontend address")
	serviceUpdateCmd.Flags().StringSliceVarP(&backends, "backends", "", []string{}, "Backend address or addresses (<IP:Port>)")

}

// cilium service update --id 1 --frontend "10.20.30.40:8081" --backends "47.243.5.64:8081,8.210.202.228:8081" --k8s-node-port
func updateService(cmd *cobra.Command, args []string) {
	id := int64(idU)
	fa, faIP := parseFrontendAddress(frontend)

	var spec *models.ServiceSpec
	svc, err := client.GetServiceID(id)
	switch {
	case err == nil && (svc.Status == nil || svc.Status.Realized == nil):
		Fatalf("Cannot update service %d: empty state", id)

	case err == nil:
		spec = svc.Status.Realized
		fmt.Printf("Updating existing service with id '%v'\n", id)

	default:
		spec = &models.ServiceSpec{ID: id}
		fmt.Printf("Creating new service with id '%v'\n", id)
	}

	// This can happen when we create a new service or when the service returned
	// to us has no flags set
	if spec.Flags == nil {
		spec.Flags = &models.ServiceSpecFlags{}
	}

	if boolToInt(k8sExternalIPs)+boolToInt(k8sNodePort)+boolToInt(k8sHostPort)+boolToInt(k8sLoadBalancer)+boolToInt(localRedirect) > 1 {
		Fatalf("Can only set one of --k8s-external, --k8s-node-port, --k8s-load-balancer, --k8s-host-port, --local-redirect for a service")
	} else if k8sExternalIPs {
		spec.Flags = &models.ServiceSpecFlags{Type: models.ServiceSpecFlagsTypeExternalIPs}
	} else if k8sNodePort {
		spec.Flags = &models.ServiceSpecFlags{Type: models.ServiceSpecFlagsTypeNodePort}
	} else if k8sLoadBalancer {
		spec.Flags = &models.ServiceSpecFlags{Type: models.ServiceSpecFlagsTypeLoadBalancer}
	} else if k8sHostPort {
		spec.Flags = &models.ServiceSpecFlags{Type: models.ServiceSpecFlagsTypeHostPort}
	} else if localRedirect {
		spec.Flags = &models.ServiceSpecFlags{Type: models.ServiceSpecFlagsTypeLocalRedirect}
	} else {
		spec.Flags = &models.ServiceSpecFlags{Type: models.ServiceSpecFlagsTypeClusterIP}
	}

	if strings.ToLower(k8sTrafficPolicy) == "local" {
		spec.Flags.TrafficPolicy = models.ServiceSpecFlagsTrafficPolicyLocal
	} else {
		spec.Flags.TrafficPolicy = models.ServiceSpecFlagsTrafficPolicyCluster
	}

	spec.FrontendAddress = fa

	if len(backends) == 0 {
		fmt.Printf("Reading backend list from stdin...\n")

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			backends = append(backends, scanner.Text())
		}
	}

	spec.BackendAddresses = nil
	for _, backend := range backends {
		beAddr, err := net.ResolveTCPAddr("tcp", backend)
		if err != nil {
			Fatalf("Cannot parse backend address \"%s\": %s", backend, err)
		}

		// Backend ID will be set by the daemon
		be := loadbalancer.NewBackend(0, loadbalancer.TCP, beAddr.IP, uint16(beAddr.Port))

		if be.IsIPv6() && faIP.To4() != nil {
			Fatalf("Address mismatch between frontend and backend %s", backend)
		}

		if fa.Port == 0 && beAddr.Port != 0 {
			Fatalf("L4 backend found (%v) with L3 frontend", beAddr)
		}

		ba := be.GetBackendModel()
		spec.BackendAddresses = append(spec.BackendAddresses, ba)
	}

	if created, err := client.PutServiceID(id, spec); err != nil {
		Fatalf("Cannot add/update service: %s", err)
	} else if created {
		fmt.Printf("Added service with %d backends\n", len(spec.BackendAddresses))
	} else {
		fmt.Printf("Updated service with %d backends\n", len(spec.BackendAddresses))
	}
}
