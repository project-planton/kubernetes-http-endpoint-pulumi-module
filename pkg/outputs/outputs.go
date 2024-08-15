package outputs

import (
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/kuberneteshttpendpoint"
	"github.com/plantoncloud/stack-job-runner-golang-sdk/pkg/automationapi/autoapistackoutput"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

const (
	Namespace = "namespace"
)

func PulumiOutputsToStackOutputsConverter(pulumiOutputs auto.OutputMap,
	input *kuberneteshttpendpoint.KubernetesHttpEndpointStackInput) *kuberneteshttpendpoint.KubernetesHttpEndpointStackOutputs {
	return &kuberneteshttpendpoint.KubernetesHttpEndpointStackOutputs{
		Namespace: autoapistackoutput.GetVal(pulumiOutputs, Namespace),
	}
}
