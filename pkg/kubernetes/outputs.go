package kubernetes

import (
	cepv1containerstack "buf.build/gen/go/plantoncloud/planton-cloud-apis/protocolbuffers/go/cloud/planton/apis/v1/code2cloud/deploy/endpoint/custom/stack/kubernetes"
	"buf.build/gen/go/plantoncloud/planton-cloud-apis/protocolbuffers/go/cloud/planton/apis/v1/stack/rpc/enums"
	"context"
	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/pulumi-stack-runner-go-sdk/pkg/org"
	"github.com/plantoncloud-inc/pulumi-stack-runner-go-sdk/pkg/stack/output/backend"
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
