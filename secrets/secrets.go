package secrets

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Secrets struct {
	ConfigName string `json:"ConfigName"`
	Records    []struct {
		Key   string `json:"Key"`
		Value string `json:"Value"`
	} `json:"Records"`
}

// InitializeFromEnvironment - accepts environment variable name that holds either a path to a file or a path to a GCP Secrets entry
func InitializeFromEnvironment(ev string) (*Secrets, error) {
	return Initialize(os.Getenv(ev))
}

// Initialize - accepts either a path to a file or a path to a GCP Secrets entry
func Initialize(location string) (*Secrets, error) {
	if len(location) == 0 {
		return nil, errors.New("secrets location not specified")
	}
	b, e := getSecretData(location)
	if e != nil {
		return nil, e
	}
	return Parse(b)
}

func (s Secrets) GetName() string {
	return s.ConfigName
}

func (s Secrets) GetString(key string) string {
	for _, r := range s.Records {
		if r.Key == key {
			return r.Value
		}
	}
	return ``
}

func (s Secrets) KeyExists(key string) bool {
	for _, r := range s.Records {
		if r.Key == key {
			return true
		}
	}
	return false
}

func (s Secrets) GetBool(key string) bool {
	v := s.GetString(key)
	return len(v) > 0 && strings.ToLower(v) == `true`
}

func (s Secrets) GetInt(key string) int {
	v := s.GetString(key)
	if len(v) > 0 {
		i, e := strconv.Atoi(v)
		if e == nil {
			return i
		}
	}
	return 0
}

// GetFile - assumes value contains location of either a file or GCP secrets entry
func (s Secrets) GetFile(key string) ([]byte, error) {
	return getSecretData(s.GetString(key))
}

// Parse - attempts to unmarshall the data into Secrets struct
func Parse(data []byte) (*Secrets, error) {
	if len(data) == 0 {
		return nil, errors.New(`no data passed to parse`)
	}
	s := new(Secrets)
	e2 := json.Unmarshal(data, &s)
	if e2 != nil {
		return nil, e2
	}
	return s, nil
}

// getGCPSecretsData - read an entry from GCP Secrets service
func getGCPSecretsData(name string) ([]byte, error) {
	// Create the client.
	ctx := context.Background()
	client, e1 := secretmanager.NewClient(ctx)
	if e1 != nil {
		return nil, fmt.Errorf("failed to create secretmanager client: %v", e1)
	}
	req := &secretmanagerpb.AccessSecretVersionRequest{Name: name}
	result, e2 := client.AccessSecretVersion(ctx, req)
	if e2 == nil && result != nil {
		return result.Payload.Data, e2
	}
	return []byte(``), e2
}

// getSecretData - returns file from specified path / cloud secrets location
// note this file is not expected to be in json format and will not be parsed by this routine
func getSecretData(fname string) ([]byte, error) {
	if len(fname) == 0 {
		return []byte(``), errors.New(`blank filename passed to secrets.GetFile`)
	}
	if strings.HasPrefix(fname, `projects/`) {
		return getGCPSecretsData(fname)
	}
	// else just read from file
	return ioutil.ReadFile(fname)
}
