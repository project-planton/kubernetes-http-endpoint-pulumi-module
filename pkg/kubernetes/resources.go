package kubernetes

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/custom-endpoint-pulumi-blueprint/pkg/kubernetes/cert"
	"github.com/plantoncloud-inc/custom-endpoint-pulumi-blueprint/pkg/kubernetes/clusterissuer"
	"github.com/plantoncloud-inc/custom-endpoint-pulumi-blueprint/pkg/kubernetes/gateway"
	"github.com/plantoncloud-inc/custom-endpoint-pulumi-blueprint/pkg/kubernetes/virtualservice"
	"github.com/plantoncloud-inc/go-commons/network/dns/zone"
	pulumikubernetesprovider "github.com/plantoncloud-inc/pulumi-stack-runner-go-sdk/pkg/automation/provider/kubernetes"
	code2cloudv1deploycepstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/customendpoint/stack/kubernetes/model"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ResourceStack struct {
	WorkspaceDir     string
	Input            *code2cloudv1deploycepstackk8smodel.CustomEndpointKubernetesStackInput
	AwsLabels        map[string]string
	KubernetesLabels map[string]string
}

func (s *ResourceStack) Resources(ctx *pulumi.Context) error {
	clusterIssuerName := GetClusterIssuerName(s.Input.ResourceInput.CustomEndpoint.Metadata.Name)
	kubernetesProvider, err := pulumikubernetesprovider.GetWithStackCredentials(ctx, s.Input.CredentialsInput.Kubernetes)
	if err != nil {
		return errors.Wrap(err, "failed to setup kubernetes provider")
	}
	if err := clusterissuer.Resources(ctx, &clusterissuer.Input{
		Workspace:                        s.WorkspaceDir,
		DnsZoneGcpProjectId:              s.Input.ResourceInput.CustomEndpoint.Spec.DnsZoneGcpProjectId,
		KubernetesProvider:               kubernetesProvider,
		LetsEncryptDns01GcpDnsIssuerName: clusterIssuerName,
	}); err != nil {
		return errors.Wrapf(err, "failed to add cluster issuer resources")
	}
	if err := addContainerResources(ctx, kubernetesProvider, s.Input.ResourceInput,
		s.KubernetesLabels, s.WorkspaceDir, clusterIssuerName); err != nil {
		return errors.Wrap(err, "failed to add external domain container resources")
	}
	return nil
}

func addContainerResources(ctx *pulumi.Context, kubernetesProvider *pulumikubernetes.Provider,
	stackResourceInput *code2cloudv1deploycepstackk8smodel.CustomEndpointKubernetesStackResourceInput,
	labels map[string]string, workspace, clusterIssuerName string) error {

	if stackResourceInput.CustomEndpoint.Spec.IsTlsEnabled {
		if err := cert.Resources(ctx, &cert.Input{
			KubernetesProvider: kubernetesProvider,
			Labels:             labels,
			EndpointDomainName: stackResourceInput.CustomEndpoint.Metadata.Name,
			Workspace:          workspace,
			ClusterIssuerName:  clusterIssuerName,
		}); err != nil {
			return errors.Wrap(err, "	failed to add cert resources")
		}
	}
	if err := gateway.Resources(ctx, &gateway.Input{
		KubernetesProvider: kubernetesProvider,
		Labels:             labels,
		EndpointDomainName: stackResourceInput.CustomEndpoint.Metadata.Name,
		Workspace:          workspace,
		IsTlsEnabled:       stackResourceInput.CustomEndpoint.Spec.IsTlsEnabled,
	}); err != nil {
		return errors.Wrap(err, "failed to add ingress gateway resources")
	}
	//virtual service resource is not required when no routes are configured for custom endpoint.
	if len(stackResourceInput.CustomEndpoint.Spec.Routes) == 0 {
		return nil
	}
	err := virtualservice.Resources(ctx, &virtualservice.Input{
		KubernetesProvider:   kubernetesProvider,
		WorkspaceDir:         workspace,
		Labels:               labels,
		EndpointDomainName:   stackResourceInput.CustomEndpoint.Metadata.Name,
		BackendEnvironmentId: stackResourceInput.CustomEndpoint.Spec.BackendEnvironmentId,
		IsGrpcWebCompatible:  stackResourceInput.CustomEndpoint.Spec.IsGrpcWebCompatible,
		Routes:               stackResourceInput.CustomEndpoint.Spec.Routes,
	})
	if err != nil {
		return errors.Wrap(err, "failed to add virtual service resources")
	}
	return nil
}

// GetClusterIssuerName returns the name of the cluster issuer to be used for all the certificate names
func GetClusterIssuerName(endpointDomainName string) string {
	return zone.GetZoneName(endpointDomainName)
}
