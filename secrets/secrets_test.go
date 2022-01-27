package secrets

import (
	"io/ioutil"
	"os"
	"testing"
)

const sfile1 = "./stest1.json"
const sfile2 = "./stest2.json"
const secretData1 = `{
  		"ConfigName": "testing",
  		"Records": [{ "Key": "StringKey", "Value": "aString" }, 
					{ "Key": "BoolKey", "Value": "true"},
					{ "Key": "FileKey", "Value": "./stest2.json"},
					{ "Key": "IntKey", "Value": "42" }	]}`

// gcp cloud secrets - identify a secret that contains same data as 'secretData2' below
const secretData2 = `{
  		"ConfigName": "testing",
  		"Records": [{ "Key": "StringKey", "Value": "aString" }, 
					{ "Key": "BoolKey", "Value": "true"},
					{ "Key": "IntKey", "Value": "42" }	]}`

func prepareSecretFile() {
	_ = ioutil.WriteFile(sfile1, []byte(secretData1), 0644)
	_ = ioutil.WriteFile(sfile2, []byte(secretData2), 0644)
}

func TestSecretsFromFile(t *testing.T) {
	prepareSecretFile()
	s, e := Initialize(sfile1)
	if e != nil {
		t.Error(e)
		return
	}
	verifySecrets(s, t)
	_ = os.Remove(sfile1)
	_ = os.Remove(sfile2)
}

// you will need to load the text from secretData1 (above) into a GCP Secret and provide the URL here for this test to work
const gcpPath = `projects/823447285116/secrets/test/versions/1`

func TestSecretsFromGCP(t *testing.T) {
	s, e := Initialize(gcpPath)
	if e != nil {
		t.Error(e)
		return
	}
	verifySecrets(s, t)
}

const fkey = `FileKey`

func verifySecrets(s *Secrets, t *testing.T) {
	if s.GetName() != `testing` {
		t.Error(`unexpected secret name`)
	}
	if s.KeyExists(`babble`) {
		t.Error(`Key should not exist`)
	}
	k1 := s.GetString(`StringKey`)
	if k1 != `aString` {
		t.Error(`bad string key`)
	}
	k2 := s.GetInt(`IntKey`)
	if k2 != 42 {
		t.Error(`bad int key`)
	}
	k3 := s.GetBool(`BoolKey`)
	if !k3 {
		t.Error(`bad bool key`)
	}
	k4 := s.GetInt(`BoolKey`)
	if k4 != 0 {
		t.Error(`bad int key`)
	}
	// test the nested secrets file feature, if it exists
	if s.KeyExists(fkey) {
		b, e := s.GetFile(fkey)
		if e != nil {
			t.Error(e)
		}
		s2, e2 := Parse(b)
		if e2 != nil {
			t.Error(e)
		}
		if s2 == nil || !s2.KeyExists(`StringKey`) {
			t.Error(`missing expected string key`)
		}
	}
}
