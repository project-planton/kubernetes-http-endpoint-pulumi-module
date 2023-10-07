package clusterissuer

import (
	"fmt"
	v12 "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	v1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	v14 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	"github.com/plantoncloud-inc/stack-runner-service/pkg/letsencrypt"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	k8sapimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
)

const (
	LetsEncryptClusterIssuerSecretName = "letsencrypt-production"
)

type Input struct {
	Workspace                        string
	DnsZoneGcpProjectId              string
	KubernetesProvider               *pulumikubernetes.Provider
	LetsEncryptDns01GcpDnsIssuerName string
}

// Resources adds cluster-issuer to be used for provisioning the certificates for the endpoint domain
func Resources(ctx *pulumi.Context, input *Input) error {
	issuerObject := buildClusterIssuerObject(input)
	resourceName := fmt.Sprintf("cluster-issuer-%s", issuerObject.Name)
	manifestPath := filepath.Join(input.Workspace, fmt.Sprintf("%s.yaml", resourceName))
	if err := manifest.Create(manifestPath, issuerObject); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", manifestPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, resourceName,
		&pulumik8syaml.ConfigFileArgs{File: manifestPath}, pulumi.Provider(input.KubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create cluster-issuer manifest")
	}
	return nil
}

func buildClusterIssuerObject(input *Input) *v1.ClusterIssuer {
	return &v1.ClusterIssuer{
		TypeMeta: k8sapimachineryv1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "ClusterIssuer",
		},
		ObjectMeta: k8sapimachineryv1.ObjectMeta{
			Name: input.LetsEncryptDns01GcpDnsIssuerName,
		},
		Spec: v1.IssuerSpec{
			IssuerConfig: v1.IssuerConfig{
				ACME: &v12.ACMEIssuer{
					PreferredChain: "",
					PrivateKey: v14.SecretKeySelector{
						LocalObjectReference: v14.LocalObjectReference{
							Name: LetsEncryptClusterIssuerSecretName,
						},
					},
					Server: letsencrypt.Server,
					Solvers: []v12.ACMEChallengeSolver{
						{
							DNS01: &v12.ACMEChallengeSolverDNS01{
								CloudDNS: &v12.ACMEIssuerDNS01ProviderCloudDNS{
									Project: input.DnsZoneGcpProjectId,
								},
							},
						},
						{
							HTTP01: &v12.ACMEChallengeSolverHTTP01{
								Ingress: &v12.ACMEChallengeSolverHTTP01Ingress{
									Class: &([]string{"istio"}[0]),
								},
							},
						},
					},
				},
			},
		},
	}
}
