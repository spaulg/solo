package workflow

import (
	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type MetadataState struct {
	metadata map[string]string
}

func LoadMetadataState(filePath string) (*MetadataState, error) {
	var metadataState *MetadataState

	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			metadataState, err = NewMetadataState()
			if err != nil {
				return nil, err
			}

			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return nil, err
			}

			if err := metadataState.SaveToFile(filePath); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		metadataState, err = LoadMetadataStateFromFile(filePath)
		if err != nil {
			return nil, err
		}
	}

	return metadataState, nil
}

func LoadMetadataStateFromFile(filePath string) (*MetadataState, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	metadataState := MetadataState{}
	if err := yaml.Unmarshal(data, &metadataState.metadata); err != nil {
		return nil, err
	}

	return &metadataState, nil
}

func NewMetadataState() (*MetadataState, error) {
	state := &MetadataState{
		metadata: make(map[string]string),
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	state.metadata["hostname"] = hostname

	return state, nil
}

func (t *MetadataState) SaveToFile(filePath string) error {
	data, err := yaml.Marshal(t.metadata)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func (t *MetadataState) ExportToGrpcMetadata() metadata.MD {
	return metadata.New(t.metadata)
}
