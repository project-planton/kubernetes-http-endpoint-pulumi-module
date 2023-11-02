package virtualservice

import (
	"fmt"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	"github.com/plantoncloud-inc/go-commons/network/dns/zone"
	ingressnamespace "github.com/plantoncloud-inc/kube-cluster-pulumi-blueprint/pkg/gcp/container/addon/istio/ingress/namespace"
	cepv1 "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/v1/code2cloud/deploy/customendpoint"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
	"strings"
)

type Input struct {
	KubernetesProvider   *pulumikubernetes.Provider
	WorkspaceDir         string
	Labels               map[string]string
	EndpointDomainName   string
	BackendEnvironmentId string
	IsGrpcWebCompatible  bool
	Routes               []*cepv1.CustomEndpointRoute
}

func Resources(ctx *pulumi.Context, input *Input) error {
	virtualServiceObject := buildVirtualServiceObject(input)
	if err := addVirtualService(ctx, virtualServiceObject, input.WorkspaceDir, input.KubernetesProvider); err != nil {
		return errors.Wrapf(err, "failed to add virtual service for %s endpoint domain name", input.EndpointDomainName)
	}
	return nil
}

func addVirtualService(ctx *pulumi.Context, virtualServiceObject *v1beta1.VirtualService, workspace string, provider *pulumikubernetes.Provider) error {
	resourceName := fmt.Sprintf("virtual-service-%s", virtualServiceObject.Name)
	manifestPath := filepath.Join(workspace, fmt.Sprintf("%s.yaml", resourceName))
	if err := manifest.Create(manifestPath, virtualServiceObject); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", manifestPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, resourceName, &pulumik8syaml.ConfigFileArgs{File: manifestPath}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to add virtual-service manifest")
	}
	return nil
}

/*
# for grpc compatible service definition with multiple routes
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:

	labels:
	  product: planton-pcs
	  company: planton
	name: api-dev-planton-pcs
	namespace: istio-ingress

spec:

		gateways:
		  - istio-ingress/api-dev-planton-cloud
		hosts:
		  - api.dev.planton.cloud
		http:
		- name: user
		  match:
		  - uri:
		      prefix: /pcs.company
		  route:
		  - destination:
		      host: main.platon-pcs-dev-company.svc.cluster.local
		      port:
		        number: 80
	      corsPolicy:
		    allowHeaders:
		    - x-user-agent
		    - authorization
		    - content-type
		    - x-grpc-web
		    allowOrigins:
		    - regex: .*
*/
func buildVirtualServiceObject(input *Input) *v1beta1.VirtualService {
	return &v1beta1.VirtualService{
		TypeMeta: k8smetav1.TypeMeta{
			APIVersion: "networking.istio.io/v1beta1",
			Kind:       "VirtualService",
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      zone.GetZoneName(input.EndpointDomainName),
			Namespace: ingressnamespace.Name,
		},
		Spec: networkingv1beta1.VirtualService{
			Gateways: []string{fmt.Sprintf("%s/%s", ingressnamespace.Name, zone.GetZoneName(input.EndpointDomainName))},
			Hosts:    []string{input.EndpointDomainName},
			Http:     getHttpRoutes(input),
		},
	}
}

func getHttpRoutes(input *Input) []*networkingv1beta1.HTTPRoute {
	httpRoutes := make([]*networkingv1beta1.HTTPRoute, 0)
	for _, r := range input.Routes {
		httpRoute := &networkingv1beta1.HTTPRoute{
			Name: getHttpRouteName(r.UrlPathPrefix),
			Match: []*networkingv1beta1.HTTPMatchRequest{
				{
					Uri: &networkingv1beta1.StringMatch{
						MatchType: &networkingv1beta1.StringMatch_Prefix{
							Prefix: r.UrlPathPrefix,
						},
					},
				},
			},
			Route: []*networkingv1beta1.HTTPRouteDestination{
				{
					Destination: &networkingv1beta1.Destination{
						Host: r.BackendKubernetesEndpoint,
						Port: &networkingv1beta1.PortSelector{Number: uint32(r.BackendMicroserviceServicePort)},
					},
				},
			},
		}
		if input.IsGrpcWebCompatible {
			/*
				corsPolicy:
				  allowOrigins:
					- regex: .*
				  allowMethods:
					- POST
					- GET
					- OPTIONS
					- PUT
					- DELETE
				  allowHeaders:
					- authorization
					- cache-control
					- content-transfer-encoding
					- content-type
					- grpc-timeout
					- keep-alive
					- user-agent
					- x-accept-content-transfer-encoding
					- x-accept-response-streaming
					- x-grpc-web
					- x-user-agent
				  maxAge: 1728s
				  exposeHeaders:
					- grpc-status
					- grpc-message
				  allowCredentials: true
			*/
			httpRoute.CorsPolicy = &networkingv1beta1.CorsPolicy{
				AllowOrigins: []*networkingv1beta1.StringMatch{{
					MatchType: &networkingv1beta1.StringMatch_Regex{Regex: ".*"},
				}},
				AllowMethods: []string{
					"POST",
					"GET",
					"OPTIONS",
					"PUT",
					"DELETE",
				},
				AllowHeaders: []string{
					"authorization",
					"cache-control",
					"content-transfer-encoding",
					"content-type",
					"grpc-timeout",
					"keep-alive",
					"user-agent",
					"x-accept-content-transfer-encoding",
					"x-accept-response-streaming",
					"x-grpc-web",
					"x-user-agent",
				},
				MaxAge: &duration.Duration{
					Seconds: 1728,
				},
				ExposeHeaders: []string{
					"grpc-status",
					"grpc-message",
				},
				AllowCredentials: &wrappers.BoolValue{
					Value: true,
				},
			}
		}
		httpRoutes = append(httpRoutes, httpRoute)
	}
	return httpRoutes
}

func getHttpRouteName(urlPathPrefix string) string {
	httpRouteName := strings.TrimSuffix(urlPathPrefix, "/")
	if httpRouteName == "" {
		return "default"
	}
	return strings.ReplaceAll(httpRouteName, "/", "-")
}
