package fileblockstore

// Implementation of BlockStore interface that stores the preflight and provable inputs in files.
//
// The preflight data is stored in at path `<base-dir>/<chainID>/preflight/<blockNumber>.json`
// The provable inputs are stored in a file named `<base-dir>/<chainID>/provable/<blockNumber>.json`

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
)

type FileBlockStore struct {
	baseDir string
}

func New(baseDir string) *FileBlockStore {
	return &FileBlockStore{baseDir: baseDir}
}

func (s *FileBlockStore) StorePreflightData(_ context.Context, data *blockinputs.PreflightData) error {
	path := s.preflightPath(data.ChainConfig.ChainID.Uint64(), data.Block.Number.ToInt().Uint64())
	return s.storeData(path, data)
}

func (s *FileBlockStore) LoadPreflightData(_ context.Context, chainID, blockNumber uint64) (*blockinputs.PreflightData, error) {
	path := s.preflightPath(chainID, blockNumber)
	data := &blockinputs.PreflightData{}
	if err := s.loadData(path, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *FileBlockStore) StoreProvableInputs(_ context.Context, data *blockinputs.ProvableInputs) error {
	path := s.provablePath(data.ChainConfig.ChainID.Uint64(), data.Block.Number.ToInt().Uint64())
	return s.storeData(path, data)
}

func (s *FileBlockStore) LoadProvableInputs(_ context.Context, chainID, blockNumber uint64) (*blockinputs.ProvableInputs, error) {
	path := s.provablePath(chainID, blockNumber)
	data := &blockinputs.ProvableInputs{}
	if err := s.loadData(path, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *FileBlockStore) preflightPath(chainID, blockNumber uint64) string {
	return filepath.Join(s.baseDir, fmt.Sprintf("%d", chainID), "preflight", fmt.Sprintf("%d.json", blockNumber))
}

func (s *FileBlockStore) provablePath(chainID, blockNumber uint64) string {
	return filepath.Join(s.baseDir, fmt.Sprintf("%d", chainID), "provable", fmt.Sprintf("%d.json", blockNumber))
}

func (s *FileBlockStore) storeData(path string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for file %s: %v", path, err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", path, err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(data); err != nil {
		return fmt.Errorf("failed to encode data to file %s: %v", path, err)
	}

	return nil
}

func (s *FileBlockStore) loadData(path string, data interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", path, err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(data); err != nil {
		return fmt.Errorf("failed to decode data from file %s: %v", path, err)
	}

	return nil
}
