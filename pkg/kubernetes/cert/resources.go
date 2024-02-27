package cert

import (
	"fmt"
	"path/filepath"

	v1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	commonsdnszone "github.com/plantoncloud-inc/go-commons/network/dns/zone"
	ingressnamespace "github.com/plantoncloud/kube-cluster-pulumi-blueprint/pkg/gcp/container/addon/istio/ingress/namespace"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	k8sapimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Namespace = ingressnamespace.Name
)

type Input struct {
	KubernetesProvider *pulumikubernetes.Provider
	Labels             map[string]string
	EndpointDomainName string
	Workspace          string
	ClusterIssuerName  string
}

func Resources(ctx *pulumi.Context, input *Input) error {
	certObj := buildCertObject(input.EndpointDomainName, Namespace, input.ClusterIssuerName, input.Labels)
	resourceName := fmt.Sprintf("cert-%s", certObj.Name)
	manifestPath := filepath.Join(input.Workspace, fmt.Sprintf("%s.yaml", resourceName))
	if err := manifest.Create(manifestPath, certObj); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", manifestPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, resourceName,
		&pulumik8syaml.ConfigFileArgs{File: manifestPath}, pulumi.Provider(input.KubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to add cert manifest")
	}
	return nil
}

/*
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:

	name: dev-example-com
	namespace: istio-ingress

spec:

	dnsNames:
	  - console.planton.cloud
	issuerRef:
	  kind: ClusterIssuer
	  name: console-planton-cloud
	secretName: cert-console-planton-cloud
*/
func buildCertObject(endpointDomainName, namespace, clusterIssuerName string, labels map[string]string) *v1.Certificate {
	return &v1.Certificate{
		TypeMeta: k8sapimachineryv1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "Certificate",
		},
		ObjectMeta: k8sapimachineryv1.ObjectMeta{
			Name:      commonsdnszone.GetZoneName(endpointDomainName),
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1.CertificateSpec{
			SecretName: GetCertSecretName(endpointDomainName),
			DNSNames:   []string{endpointDomainName},
			IssuerRef:  cmmeta.ObjectReference{Kind: "ClusterIssuer", Name: clusterIssuerName},
		},
	}
}

func GetCertSecretName(endpointDomainName string) string {
	return fmt.Sprintf("cert-%s", commonsdnszone.GetZoneName(endpointDomainName))
}
