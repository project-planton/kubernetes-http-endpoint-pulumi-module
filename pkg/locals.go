package pkg

import (
	"fmt"
	"github.com/plantoncloud/kubernetes-http-endpoint-pulumi-module/pkg/outputs"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/kuberneteshttpendpoint"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/apiresource/enums/apiresourcekind"
	"github.com/plantoncloud/pulumi-module-golang-commons/pkg/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"strconv"
)

type Locals struct {
	EndpointDomainName     string
	IngressCertSecretName  string
	KubernetesHttpEndpoint *kuberneteshttpendpoint.KubernetesHttpEndpoint
	KubernetesLabels       map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kuberneteshttpendpoint.KubernetesHttpEndpointStackInput) *Locals {
	locals := &Locals{}

	//assign value for the locals variable to make it available across the project
	locals.KubernetesHttpEndpoint = stackInput.ApiResource

	locals.KubernetesLabels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.Organization: locals.KubernetesHttpEndpoint.Spec.EnvironmentInfo.OrgId,
		kuberneteslabelkeys.Environment:  locals.KubernetesHttpEndpoint.Spec.EnvironmentInfo.EnvId,
		kuberneteslabelkeys.ResourceKind: apiresourcekind.ApiResourceKind_kafka_kubernetes.String(),
		kuberneteslabelkeys.ResourceId:   locals.KubernetesHttpEndpoint.Metadata.Id,
	}

	locals.EndpointDomainName = locals.KubernetesHttpEndpoint.Metadata.Name

	locals.IngressCertSecretName = fmt.Sprintf("cert-%s", locals.KubernetesHttpEndpoint.Metadata.Name)

	ctx.Export(outputs.Namespace, pulumi.String(vars.IstioIngressNamespace))

	return locals
}
