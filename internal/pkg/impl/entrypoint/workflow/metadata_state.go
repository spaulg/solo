package workflow

import (
	"os"
	"path/filepath"

	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v3"
)

type MetadataState struct {
	filePath string
	metadata map[string]string
	dirty    bool
}

func LoadMetadataState(filePath string) (*MetadataState, error) {
	var metadataState *MetadataState

	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return nil, err
			}

			metadataState, err = NewMetadataState(filePath)
			if err != nil {
				return nil, err
			}

			if err := metadataState.SaveToFile(); err != nil {
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

	metadataState := MetadataState{
		filePath: filePath,
		dirty:    false,
	}

	if err := yaml.Unmarshal(data, &metadataState.metadata); err != nil {
		return nil, err
	}

	return &metadataState, nil
}

func NewMetadataState(filePath string) (*MetadataState, error) {
	state := &MetadataState{
		filePath: filePath,
		metadata: make(map[string]string),
		dirty:    false,
	}

	return state, nil
}

func (t *MetadataState) Set(key string, value string) {
	t.metadata[key] = value
	t.dirty = true
}

func (t *MetadataState) SaveToFile() error {
	if !t.dirty {
		return nil
	}

	data, err := yaml.Marshal(t.metadata)
	if err != nil {
		return err
	}

	if err := os.WriteFile(t.filePath, data, 0600); err != nil {
		return err
	}

	t.dirty = false
	return nil
}

func (t *MetadataState) ExportToGrpcMetadata() metadata.MD {
	return metadata.New(t.metadata)
}
