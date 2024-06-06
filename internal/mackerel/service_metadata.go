package mackerel

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mackerelio/mackerel-client-go"
)

func ServiceMetadataID(serviceName, namespace string) string {
	return strings.Join([]string{serviceName, namespace}, "/")
}

func ParseServiceMetadataID(id string) (serviceName, namespace string, err error) {
	first, last, ok := strings.Cut(id, "/")
	if !ok {
		return "", "", fmt.Errorf("The ID is expected to have `<service_name>/<namespace>` format, but got: '%q'.", id)
	}
	return first, last, nil
}

type ServiceMetadataModel struct {
	ID           types.String         `tfsdk:"id"`
	ServiceName  types.String         `tfsdk:"service"`
	Namespace    types.String         `tfsdk:"namespace"`
	MetadataJSON jsontypes.Normalized `tfsdk:"metadata_json"`
}

func ReadServiceMetadata(_ context.Context, client *Client, serviceName, namespace string) (*ServiceMetadataModel, error) {
	var data ServiceMetadataModel
	data.ServiceName = types.StringValue(serviceName)
	data.Namespace = types.StringValue(namespace)
	data.ID = types.StringValue(ServiceMetadataID(serviceName, namespace))

	metadataResp, err := client.GetServiceMetaData(serviceName, namespace)
	if err != nil {
		return nil, err
	}

	data.ServiceName = types.StringValue(serviceName)
	data.Namespace = types.StringValue(namespace)
	data.MetadataJSON = jsontypes.NewNormalizedValue("")

	metadataMap, ok := metadataResp.ServiceMetaData.(map[string]any)
	if !ok || len(metadataMap) == 0 {
		return &data, nil
	}

	metadataJSON, err := json.Marshal(metadataMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	data.MetadataJSON = jsontypes.NewNormalizedValue(string(metadataJSON))
	return &data, nil
}

func (m *ServiceMetadataModel) Set(newData ServiceMetadataModel) {
	*m = newData
}

func (m *ServiceMetadataModel) CreateOrUpdateMetadata(_ context.Context, client *Client) error {
	serviceName, namespace := m.ServiceName.ValueString(), m.Namespace.ValueString()

	var metadata mackerel.ServiceMetaData

	if err := json.Unmarshal(
		[]byte(m.MetadataJSON.ValueString()), &metadata,
	); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	if err := client.PutServiceMetaData(serviceName, namespace, metadata); err != nil {
		return err
	}

	m.ID = types.StringValue(ServiceMetadataID(serviceName, namespace))
	return nil
}
