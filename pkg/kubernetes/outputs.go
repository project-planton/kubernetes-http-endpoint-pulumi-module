package kubernetes

import (
	"context"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/pulumi-stack-runner-go-sdk/pkg/org"
	"github.com/plantoncloud-inc/pulumi-stack-runner-go-sdk/pkg/stack/output/backend"
	cepv1containerstack "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/v1/code2cloud/deploy/endpoint/custom/stack/kubernetes"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/v1/stack/enums"
)

func Outputs(ctx context.Context, input *cepv1containerstack.CustomEndpointKubernetesStackInput) (
	*cepv1containerstack.CustomEndpointKubernetesStackOutputs, error) {
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

func Get(stackOutput map[string]interface{}, input *cepv1containerstack.CustomEndpointKubernetesStackInput) *cepv1containerstack.CustomEndpointKubernetesStackOutputs {
	if input.StackJob.OperationType != enums.StackOperationType_apply || stackOutput == nil {
		return &cepv1containerstack.CustomEndpointKubernetesStackOutputs{}
	}
	return &cepv1containerstack.CustomEndpointKubernetesStackOutputs{}
}
