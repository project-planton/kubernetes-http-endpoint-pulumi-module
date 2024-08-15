package pkg

import (
	"github.com/pkg/errors"
	certmanagerv1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/certmanager/certmanager/v1"
	gatewayv1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/gatewayapis/gateway/v1"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/kuberneteshttpendpoint"
	"github.com/plantoncloud/pulumi-module-golang-commons/pkg/provider/kubernetes/pulumikubernetesprovider"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ResourceStack struct {
	StackInput *kuberneteshttpendpoint.KubernetesHttpEndpointStackInput
}

func (s *ResourceStack) Resources(ctx *pulumi.Context) error {
	locals := initializeLocals(ctx, s.StackInput)

	//create kubernetes-provider from the credential in the stack-input
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesClusterCredential(ctx,
		s.StackInput.KubernetesClusterCredential, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	listenersArray := gatewayv1.GatewaySpecListenersArray{
		&gatewayv1.GatewaySpecListenersArgs{
			Name:     pulumi.String("http"),
			Hostname: pulumi.String(locals.EndpointDomainName),
			Port:     pulumi.Int(80),
			Protocol: pulumi.String("HTTP"),
			AllowedRoutes: gatewayv1.GatewaySpecListenersAllowedRoutesArgs{
				Namespaces: gatewayv1.GatewaySpecListenersAllowedRoutesNamespacesArgs{
					From: pulumi.String("All"),
				},
			},
		},
	}

	if locals.KubernetesHttpEndpoint.Spec.IsTlsEnabled {
		// Create new certificate
		_, err := certmanagerv1.NewCertificate(ctx,
			"ingress-certificate",
			&certmanagerv1.CertificateArgs{
				Metadata: metav1.ObjectMetaArgs{
					Name:      pulumi.String(locals.KubernetesHttpEndpoint.Metadata.Id),
					Namespace: pulumi.String(vars.IstioIngressNamespace),
					Labels:    pulumi.ToStringMap(locals.KubernetesLabels),
				},
				Spec: certmanagerv1.CertificateSpecArgs{
					DnsNames:   pulumi.ToStringArray([]string{locals.KubernetesHttpEndpoint.Metadata.Name}),
					SecretName: pulumi.String(locals.IngressCertSecretName),
					IssuerRef: certmanagerv1.CertificateSpecIssuerRefArgs{
						Kind: pulumi.String("ClusterIssuer"),
						Name: pulumi.String(locals.KubernetesHttpEndpoint.Spec.CertClusterIssuerName),
					},
				},
			}, pulumi.Provider(kubernetesProvider))
		if err != nil {
			return errors.Wrap(err, "error creating certificate")
		}

		listenersArray = append(listenersArray, &gatewayv1.GatewaySpecListenersArgs{
			Name:     pulumi.String("https"),
			Hostname: pulumi.String(locals.EndpointDomainName),
			Port:     pulumi.Int(443),
			Protocol: pulumi.String("HTTPS"),
			Tls: &gatewayv1.GatewaySpecListenersTlsArgs{
				Mode: pulumi.String("Terminate"),
				CertificateRefs: gatewayv1.GatewaySpecListenersTlsCertificateRefsArray{
					gatewayv1.GatewaySpecListenersTlsCertificateRefsArgs{
						Name: pulumi.String(locals.IngressCertSecretName),
					},
				},
			},
			AllowedRoutes: gatewayv1.GatewaySpecListenersAllowedRoutesArgs{
				Namespaces: gatewayv1.GatewaySpecListenersAllowedRoutesNamespacesArgs{
					From: pulumi.String("All"),
				},
			},
		})
	}

	// Create external gateway
	createdGateway, err := gatewayv1.NewGateway(ctx,
		"gateway",
		&gatewayv1.GatewayArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String(locals.KubernetesHttpEndpoint.Metadata.Id),
				// All gateway resources should be created in the ingress deployment namespace
				Namespace: pulumi.String(vars.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(locals.KubernetesLabels),
			},
			Spec: gatewayv1.GatewaySpecArgs{
				GatewayClassName: pulumi.String(vars.GatewayIngressClassName),
				Addresses: pulumi.Array{
					pulumi.Map{
						"type":  pulumi.String("Hostname"),
						"value": pulumi.String(vars.GatewayExternalLoadBalancerServiceHostname),
					},
				},
				Listeners: listenersArray,
			},
		}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "error creating gateway")
	}

	httpRulesArray := gatewayv1.HTTPRouteSpecRulesArray{}

	//build http-rules based on the routes configured in the input
	for _, routingRule := range locals.KubernetesHttpEndpoint.Spec.RoutingRules {
		httpRulesArray = append(httpRulesArray,
			gatewayv1.HTTPRouteSpecRulesArgs{
				Matches: gatewayv1.HTTPRouteSpecRulesMatchesArray{
					gatewayv1.HTTPRouteSpecRulesMatchesArgs{
						Path: gatewayv1.HTTPRouteSpecRulesMatchesPathArgs{
							Type:  pulumi.String("PathPrefix"),
							Value: pulumi.String(routingRule.UrlPathPrefix),
						},
					},
				},
				BackendRefs: gatewayv1.HTTPRouteSpecRulesBackendRefsArray{
					gatewayv1.HTTPRouteSpecRulesBackendRefsArgs{
						Name:      pulumi.String(routingRule.BackendService.Name),
						Namespace: pulumi.String(routingRule.BackendService.Namespace),
						Port:      pulumi.Int(routingRule.BackendService.Port),
					},
				},
			})
	}

	// Create HTTP route with routing rules
	_, err = gatewayv1.NewHTTPRoute(ctx,
		"http-route",
		&gatewayv1.HTTPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.KubernetesHttpEndpoint.Metadata.Id),
				Namespace: pulumi.String(vars.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(locals.KubernetesLabels),
			},
			Spec: gatewayv1.HTTPRouteSpecArgs{
				Hostnames: pulumi.StringArray{pulumi.String(locals.EndpointDomainName)},
				ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
					gatewayv1.HTTPRouteSpecParentRefsArgs{
						Name:      pulumi.Sprintf("%s", createdGateway.Metadata.Name()),
						Namespace: createdGateway.Metadata.Namespace(),
					},
				},
				Rules: httpRulesArray,
			},
		}, pulumi.Parent(createdGateway))

	if err != nil {
		return errors.Wrap(err, "error creating HTTP route")
	}
	return nil
}
