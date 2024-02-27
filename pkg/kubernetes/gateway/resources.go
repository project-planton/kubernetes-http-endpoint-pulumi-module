package gateway

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	"github.com/plantoncloud-inc/go-commons/network/dns/zone"
	"github.com/plantoncloud/custom-endpoint-pulumi-blueprint/pkg/kubernetes/cert"
	"github.com/plantoncloud/kube-cluster-pulumi-blueprint/pkg/gcp/container/addon/istio/ingress/controller"
	ingressnamespace "github.com/plantoncloud/kube-cluster-pulumi-blueprint/pkg/gcp/container/addon/istio/ingress/namespace"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"path/filepath"

	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Namespace = ingressnamespace.Name
)

type Input struct {
	KubernetesProvider *pulumikubernetes.Provider
	Labels             map[string]string
	EndpointDomainName string
	Workspace          string
	IsTlsEnabled       bool
}

func Resources(ctx *pulumi.Context, input *Input) error {
	gatewayObject := buildGatewayObject(input.EndpointDomainName, Namespace, input.Labels, input.IsTlsEnabled)
	resourceName := fmt.Sprintf("gateway-%s", zone.GetZoneName(input.EndpointDomainName))
	manifestPath := filepath.Join(input.Workspace, fmt.Sprintf("%s.yaml", resourceName))
	if err := manifest.Create(manifestPath, gatewayObject); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", manifestPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, resourceName, &pulumik8syaml.ConfigFileArgs{File: manifestPath},
		pulumi.Provider(input.KubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to add ingress-gateway manifest")
	}
	return nil
}

/*
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:

	name: dev-k8sk8s-com
	namespace: istio-ingress

spec:

	selector:
	  app: istio-ingress
	  istio: ingress
	servers:
	  - port:
	      number: 443
	      name: https
	      protocol: HTTPS
	    hosts:
	      - "*.dev.k8sk8s.com"
	    tls:
	      mode: SIMPLE
	      credentialName: cert-dev-pcsapps-com
	  - port:
	      number: 80
	      name: http
	      protocol: HTTP
	    hosts:
	      - "*.dev.k8sk8s.com"
	    tls:
	      httpsRedirect: true
*/
func buildGatewayObject(endpointDomainName, namespace string, labels map[string]string, isTlsEnabled bool) *v1beta1.Gateway {
	return &v1beta1.Gateway{
		TypeMeta: k8smetav1.TypeMeta{
			APIVersion: "networking.istio.io/v1beta1",
			Kind:       "Gateway",
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      zone.GetZoneName(endpointDomainName),
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: networkingv1beta1.Gateway{
			Selector: controller.SelectorLabels,
			Servers:  getGatewaySpecServers(endpointDomainName, isTlsEnabled),
		},
	}
}

func getGatewaySpecServers(endpointDomainName string, isTlsEnabled bool) []*networkingv1beta1.Server {
	servers := make([]*networkingv1beta1.Server, 0)
	servers = append(servers, &networkingv1beta1.Server{
		Port: &networkingv1beta1.Port{
			Number:   80,
			Protocol: "HTTP",
			Name:     "http",
		},
		Hosts: []string{endpointDomainName},
		Tls: &networkingv1beta1.ServerTLSSettings{
			HttpsRedirect: isTlsEnabled,
		},
		Name: "http",
	})
	if !isTlsEnabled {
		return servers
	}
	servers = append(servers, &networkingv1beta1.Server{
		Port: &networkingv1beta1.Port{
			Number:   443,
			Protocol: "HTTPS",
			Name:     "https",
		},
		Hosts: []string{endpointDomainName},
		Tls: &networkingv1beta1.ServerTLSSettings{
			Mode:           networkingv1beta1.ServerTLSSettings_SIMPLE,
			CredentialName: cert.GetCertSecretName(endpointDomainName),
		},
		Name: "https",
	})
	return servers
}
