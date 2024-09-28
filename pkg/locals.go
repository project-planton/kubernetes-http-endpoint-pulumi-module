package pkg

import (
	"fmt"
	"github.com/plantoncloud/kubernetes-http-endpoint-pulumi-module/pkg/outputs"
	"github.com/plantoncloud/project-planton/apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/kuberneteshttpendpoint"
	"github.com/plantoncloud/project-planton/apis/zzgo/cloud/planton/apis/commons/apiresource/enums/apiresourcekind"
	"github.com/plantoncloud/pulumi-module-golang-commons/pkg/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"strconv"
)

type Locals struct {
	EndpointDomainName     string
	IngressCertSecretName  string
	KubernetesHttpEndpoint *kuberneteshttpendpoint.KubernetesHttpEndpoint
	Labels                 map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kuberneteshttpendpoint.KubernetesHttpEndpointStackInput) *Locals {
	locals := &Locals{}

	//assign value for the locals variable to make it available across the project
	locals.KubernetesHttpEndpoint = stackInput.Target

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Environment:  stackInput.Target.Spec.EnvironmentInfo.EnvId,
		kuberneteslabelkeys.Organization: stackInput.Target.Spec.EnvironmentInfo.OrgId,
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceId:   stackInput.Target.Metadata.Id,
		kuberneteslabelkeys.ResourceKind: apiresourcekind.ApiResourceKind_kubernetes_http_endpoint.String(),
	}

	locals.EndpointDomainName = locals.KubernetesHttpEndpoint.Metadata.Name

	locals.IngressCertSecretName = fmt.Sprintf("cert-%s", locals.KubernetesHttpEndpoint.Metadata.Name)

	ctx.Export(outputs.Namespace, pulumi.String(vars.IstioIngressNamespace))

	return locals
}
