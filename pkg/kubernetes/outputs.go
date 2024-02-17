package kubernetes

import (
	"context"

	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/v1/stack/job/enums/operationtype"

	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/pulumi-stack-runner-go-sdk/pkg/org"
	"github.com/plantoncloud-inc/pulumi-stack-runner-go-sdk/pkg/stack/output/backend"
	code2cloudv1deploycepstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/v1/code2cloud/deploy/customendpoint/stack/kubernetes/model"
)

func Outputs(ctx context.Context, input *code2cloudv1deploycepstackk8smodel.CustomEndpointKubernetesStackInput) (
	*code2cloudv1deploycepstackk8smodel.CustomEndpointKubernetesStackOutputs, error) {
	pulumiOrgName, err := org.GetOrgName()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pulumi org name")
	}
	stackOutput, err := backend.StackOutput(pulumiOrgName, input.StackJob)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get stack output")
	}
	return Get(stackOutput, input), nil
}

func Get(stackOutput map[string]interface{}, input *code2cloudv1deploycepstackk8smodel.CustomEndpointKubernetesStackInput) *code2cloudv1deploycepstackk8smodel.CustomEndpointKubernetesStackOutputs {
	if input.StackJob.Spec.OperationType != operationtype.StackJobOperationType_apply || stackOutput == nil {
		return &code2cloudv1deploycepstackk8smodel.CustomEndpointKubernetesStackOutputs{}
	}
	return &code2cloudv1deploycepstackk8smodel.CustomEndpointKubernetesStackOutputs{}
}
