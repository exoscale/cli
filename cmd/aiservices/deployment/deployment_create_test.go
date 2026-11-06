package deployment

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/exoapi/mocks"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
)

func TestDeploymentCreate(t *testing.T) {
	const (
		modelID  = "4a1e7cd2-9f0b-4c3a-8e15-6b2d7f90a1c3"
		endpoint = v3.Endpoint("https://dk2")
		zone     = v3.ZoneName("ch-dk-2")
	)
	errAPI := errors.New("api failure")
	op := &v3.Operation{ID: "op-deploy-create"}

	// expectCreate expects the deployment to be created with exactly req.
	expectCreate := func(c *mocks.MockClient, req v3.CreateDeploymentRequest) {
		c.EXPECT().CreateDeployment(mock.Anything, req).Return(op, nil).Once()
		c.EXPECT().Wait(mock.Anything, op, v3.OperationStateSuccess).Return(op, nil).Once()
	}
	// helpVersion matches the option selecting the given inference engine version.
	helpVersion := func(version string) any {
		return mock.MatchedBy(func(opt v3.GetInferenceEngineHelpOpt) bool {
			query := url.Values{}
			opt(query)
			return query.Get("version") == version
		})
	}
	help := &v3.GetInferenceEngineHelpResponse{
		Parameters: []v3.InferenceEngineParameterEntry{
			{
				Name:        "config-format",
				Flags:       []string{"--config-format"},
				Type:        "enum",
				Default:     "auto",
				Section:     "ModelConfig",
				Description: "The format of the model config to load.",
			},
			{
				Name:        "max-model-len",
				Flags:       []string{"--max-model-len"},
				Type:        "integer",
				Default:     "None",
				Section:     "ModelConfig",
				Description: "Model context length.",
			},
		},
	}

	tests := []struct {
		name   string
		cmd    DeploymentCreateCmd
		client func(*mocks.MockClient)
		// err is a part of the expected error message, none is expected when empty.
		err string
	}{
		{
			name: "missing name",
			cmd:  DeploymentCreateCmd{GPUType: "gpua5000", GPUCount: 1, ModelName: "m1"},
			err:  "NAME is required",
		},
		{
			name: "missing gpu type",
			cmd:  DeploymentCreateCmd{Name: "dep", GPUCount: 1, ModelName: "m1"},
			err:  "--gpu-type and --gpu-count are required",
		},
		{
			name: "missing gpu count",
			cmd:  DeploymentCreateCmd{Name: "dep", GPUType: "gpua5000", ModelName: "m1"},
			err:  "--gpu-type and --gpu-count are required",
		},
		{
			name: "missing model",
			cmd:  DeploymentCreateCmd{Name: "dep", GPUType: "gpua5000", GPUCount: 1},
			err:  "--model-id or --model-name is required",
		},
		{
			name: "invalid model id",
			cmd:  DeploymentCreateCmd{Name: "dep", GPUType: "gpua5000", GPUCount: 1, ModelID: "bad"},
			err:  "invalid --model-id",
		},
		{
			name: "with model name",
			cmd:  DeploymentCreateCmd{Name: "dep", GPUType: "gpua5000", GPUCount: 1, Replicas: 1, ModelName: "m1"},
			client: func(c *mocks.MockClient) {
				expectCreate(c, v3.CreateDeploymentRequest{
					Name: "dep", GpuType: "gpua5000", GpuCount: 1, Replicas: 1,
					Model: &v3.ModelRef{Name: "m1"},
				})
			},
		},
		{
			name: "with model id",
			cmd:  DeploymentCreateCmd{Name: "dep", GPUType: "gpua5000", GPUCount: 1, Replicas: 1, ModelID: modelID},
			client: func(c *mocks.MockClient) {
				expectCreate(c, v3.CreateDeploymentRequest{
					Name: "dep", GpuType: "gpua5000", GpuCount: 1, Replicas: 1,
					Model: &v3.ModelRef{ID: v3.UUID(modelID)},
				})
			},
		},
		{
			name: "with inference engine parameters",
			cmd: DeploymentCreateCmd{
				Name: "dep", GPUType: "gpua5000", GPUCount: 1, Replicas: 1, ModelName: "m1",
				InferenceEngineParameters: "--gpu-memory-usage=0.8 --max-tokens=4096 --disable-x-feature",
			},
			client: func(c *mocks.MockClient) {
				expectCreate(c, v3.CreateDeploymentRequest{
					Name: "dep", GpuType: "gpua5000", GpuCount: 1, Replicas: 1,
					Model:                     &v3.ModelRef{Name: "m1"},
					InferenceEngineParameters: []string{"--gpu-memory-usage=0.8", "--max-tokens=4096", "--disable-x-feature"},
				})
			},
		},
		{
			name: "with inference engine version",
			cmd: DeploymentCreateCmd{
				Name: "dep", GPUType: "gpua5000", GPUCount: 1, Replicas: 1, ModelName: "m1",
				InferenceEngineVersion: "0.15.1",
			},
			client: func(c *mocks.MockClient) {
				expectCreate(c, v3.CreateDeploymentRequest{
					Name: "dep", GpuType: "gpua5000", GpuCount: 1, Replicas: 1,
					Model:                  &v3.ModelRef{Name: "m1"},
					InferenceEngineVersion: "0.15.1",
				})
			},
		},
		{
			name: "creates in the requested zone",
			cmd:  DeploymentCreateCmd{Name: "dep", GPUType: "gpua5000", GPUCount: 1, Replicas: 1, ModelName: "m1", Zone: zone},
			client: func(c *mocks.MockClient) {
				// The create and the wait are only expected on the zonal client.
				zoned := mocks.NewMockClient(t)
				c.EXPECT().GetZoneAPIEndpoint(mock.Anything, zone).Return(endpoint, nil).Once()
				c.EXPECT().WithEndpoint(endpoint).Return(zoned).Once()
				expectCreate(zoned, v3.CreateDeploymentRequest{
					Name: "dep", GpuType: "gpua5000", GpuCount: 1, Replicas: 1,
					Model: &v3.ModelRef{Name: "m1"},
				})
			},
		},
		{
			name: "zone lookup failure",
			cmd:  DeploymentCreateCmd{Name: "dep", GPUType: "gpua5000", GPUCount: 1, ModelName: "m1", Zone: zone},
			client: func(c *mocks.MockClient) {
				c.EXPECT().GetZoneAPIEndpoint(mock.Anything, zone).Return("", errAPI).Once()
			},
			err: errAPI.Error(),
		},
		{
			name: "create failure does not wait",
			cmd:  DeploymentCreateCmd{Name: "dep", GPUType: "gpua5000", GPUCount: 1, ModelName: "m1"},
			client: func(c *mocks.MockClient) {
				c.EXPECT().CreateDeployment(mock.Anything, mock.Anything).Return(nil, errAPI).Once()
			},
			err: errAPI.Error(),
		},
		{
			name: "operation failure",
			cmd:  DeploymentCreateCmd{Name: "dep", GPUType: "gpua5000", GPUCount: 1, ModelName: "m1"},
			client: func(c *mocks.MockClient) {
				c.EXPECT().CreateDeployment(mock.Anything, mock.Anything).Return(op, nil).Once()
				c.EXPECT().Wait(mock.Anything, op, v3.OperationStateSuccess).Return(nil, errAPI).Once()
			},
			err: errAPI.Error(),
		},
		{
			name: "help needs no other argument",
			cmd:  DeploymentCreateCmd{InferenceEngineHelp: true},
			client: func(c *mocks.MockClient) {
				c.EXPECT().GetInferenceEngineHelp(mock.Anything).Return(help, nil).Once()
			},
		},
		{
			name: "help does not create the named deployment",
			cmd:  DeploymentCreateCmd{InferenceEngineHelp: true, Name: "dep"},
			client: func(c *mocks.MockClient) {
				c.EXPECT().GetInferenceEngineHelp(mock.Anything).Return(help, nil).Once()
			},
		},
		{
			name: "help for a version",
			cmd:  DeploymentCreateCmd{InferenceEngineHelp: true, InferenceEngineVersion: "0.15.1"},
			client: func(c *mocks.MockClient) {
				c.EXPECT().GetInferenceEngineHelp(mock.Anything, helpVersion("0.15.1")).Return(help, nil).Once()
			},
		},
		{
			name: "help failure",
			cmd:  DeploymentCreateCmd{InferenceEngineHelp: true},
			client: func(c *mocks.MockClient) {
				c.EXPECT().GetInferenceEngineHelp(mock.Anything).Return(nil, errAPI).Once()
			},
			err: errAPI.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mocks.NewMockClient(t)
			if tt.client != nil {
				tt.client(client)
			}
			testutils.UseClient(t, client)

			cmd := tt.cmd
			cmd.CliCommandSettings = exocmd.DefaultCLICmdSettings()

			err := cmd.CmdRun(nil, nil)
			if tt.err == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.err)
			}
		})
	}
}
