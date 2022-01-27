package storage

import (
	"cloud.google.com/go/storage"
	"errors"
	"fmt"
	"github.com/cambefus/gcp_go_utils/secrets"
	"github.com/cambefus/gcp_go_utils/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// helper routines
var cs *CStore
var tfc = 0

const fileContents = "this is sample content \n for the file"
const testPath = `testing/`

func setup(t *testing.T) {
	if cs == nil {
		s, e := secrets.InitializeFromEnvironment(`utilities_config`)
		if e != nil {
			t.Fatal(e)
		}
		cr, e1 := s.GetFile(`STORAGE_CREDENTIALS`)
		if e1 != nil {
			t.Fatal(e1)
		}
		tmp, e2 := NewCStore(cr, s.GetString(`CLOUD_STORAGE_BUCKET`))
		if e2 != nil {
			t.Fatal(e2)
		}
		cs = tmp
	}
}

// cloud storage will invoke rate limit errors if you perform to many calls against an
// individual file too quickly, so we need to change the file name for each test
func writeTestFiles(t *testing.T) []string {
	tfc++
	path := testPath + strconv.Itoa(tfc) + `/`
	result := []string{path + `f1.txt`, path + `f2.json`}

	if !(cs.WriteFile(result[0], fileContents) == nil) {
		t.Fatal(errors.New("unable to write to " + result[0]))
	}
	if !(cs.WriteCloudFile(result[1], []byte(fileContents), `application/json`) == nil) {
		t.Fatal(errors.New("unable to write to " + result[1]))
	}
	return result
}

func deleteTestFiles(t *testing.T, fnames []string) {
	for _, cfile := range fnames {
		err := cs.DeleteCloudFile(cfile)
		if err != nil {
			t.Error("Failed to delete "+cfile, err)
		}
	}
}

// end helper routines

func Test_NewCStoreP(t *testing.T) {
	// this function should never work in a test environment, as it will not have permissions
	csp, err := NewCStoreP("Sample")
	if err != nil {
		t.Error("NewCStoreP: ", err)
	}
	fl, e := csp.GetFiles("")
	if e == nil {
		t.Error("this should have failed, path does not exist")
	}
	if len(fl) != 0 {
		t.Error("GCP Storage: should not have found files")
	}

	csp2, _ := NewCStoreP("Sample2")
	if !(csp2.client == csp.client) {
		t.Error("Storage client should have been reused")
	}
}

func Test_FileExists(t *testing.T) {
	setup(t)
	f := writeTestFiles(t)
	if !cs.FileExists(f[0]) {
		t.Error(`cloud storage verification function failed`)
	}
	deleteTestFiles(t, f)
	if cs.FileExists(f[0]) {
		t.Error(`should not have found file`)
	}
}

func Test_getFiles(t *testing.T) {
	setup(t)
	f := writeTestFiles(t)
	fl, e := cs.GetFiles(testPath)
	if e != nil {
		t.Error(e)
	}
	if len(fl) != 2 {
		t.Error("GCP Storage: failed to find expected files")
	}
	deleteTestFiles(t, f)
}

func Test_GetFilesBFilter(t *testing.T) {
	setup(t)
	f := writeTestFiles(t)
	fcnt := 0
	var pf = func(oa *storage.ObjectAttrs) bool {
		if strings.HasSuffix(oa.Name, `txt`) {
			fcnt++
		}
		return true
	}
	e := cs.GetFilteredFiles(testPath, pf)
	if e != nil {
		t.Error(e)
	}
	if fcnt != 1 {
		t.Error("GetFilesBFilter: failed to return expected count")
	}
	deleteTestFiles(t, f)
}

func Test_copy(t *testing.T) {
	setup(t)
	fn := testPath + `cpy.txt`
	if !(cs.WriteFile(fn, fileContents) == nil) {
		t.Error(fn)
	}
	fn2 := testPath + `cpy2.txt`
	if !(cs.CopyFile(fn, cs, fn2) == nil) {
		t.Error("copy failed")
	}

	e := cs.DeleteCloudFile(fn)
	if e != nil {
		t.Error(e)
	}
	e2 := cs.DeleteCloudFile(fn2)
	if e2 != nil {
		t.Error(e2)
	}
}

func Test_getFileReader(t *testing.T) {
	setup(t)
	f := writeTestFiles(t)
	cr, _, e := cs.GetFileReader(f[0])
	if e != nil {
		t.Error(e)
	}
	defer cr.Close()
	c2, e2 := ioutil.ReadAll(cr)
	if e2 != nil {
		t.Error(e2)
	}

	if fileContents != string(c2) {
		t.Errorf(`Expected "%s" got "%s"`, fileContents, c2)
	}
	deleteTestFiles(t, f)
}

func Test_DeleteOldFiles(t *testing.T) {
	setup(t)
	writeTestFiles(t)
	i1, e := cs.DeleteOldFiles(testPath, 1)
	if e != nil {
		t.Error(e)
	}
	if i1 != 0 {
		t.Errorf("Expected 0, got %d", i1)
	}
	i2, e2 := cs.DeleteOldFiles(testPath, -1)
	if e2 != nil {
		t.Error(e2)
	}
	if i2 != 2 {
		t.Errorf("Expected 2, got %d", i2)
	}
	r, e3 := cs.GetFileInfo(testPath)
	if e3 != nil {
		t.Error(e3)
	}
	if len(r) != 0 {
		t.Error(`Expected 0 files to be left`)
	}
}

func Test_DownloadFiles(t *testing.T) {
	setup(t)
	f := writeTestFiles(t)
	const localDir = `./`
	flist, e := cs.GetFiles(testPath)
	if e != nil {
		t.Error(e)
	}
	err := cs.DownloadFiles(flist, localDir)
	if err != nil {
		t.Error(`DownloadFiles: `, err)
	}
	for _, i2 := range flist {
		fb := filepath.Base(i2)
		if !util.FileExists(localDir + fb) {
			t.Error(`DownloadFiles - failed to find local file`)
		}
		_ = os.Remove(localDir + fb)
	}
	deleteTestFiles(t, f)
}

func Test_getFilesWithSuffix(t *testing.T) {
	setup(t)
	f := writeTestFiles(t)
	defer deleteTestFiles(t, f)

	f1, e := cs.GetFiles(testPath)
	if e != nil {
		t.Error(e)
	}
	f2, e2 := cs.GetFilesWithSuffix(testPath, `.txt`)
	if e2 != nil {
		t.Error(e2)
	}
	if (len(f1) != 2) || (len(f2) != 1) {
		t.Errorf("Expected 2 & 1, got %d & %d", len(f1), len(f2))
	}
}

func Test_CreateDownloadURL(t *testing.T) {
	setup(t)
	f := writeTestFiles(t)
	r, e := cs.GetFiles(testPath)
	if e != nil {
		t.Error(e)
	}
	defer deleteTestFiles(t, f)
	if len(r) == 0 {
		t.Error(`No files found in bucket`)
	} else {
		u, e1 := cs.CreateDownloadURL(5, r[1])
		if e1 != nil {
			t.Errorf("Got error %v", e1)
		} else {
			fmt.Println(u)
		}
	}
}

func Test_GetFileInfo(t *testing.T) {
	setup(t)
	f := writeTestFiles(t)
	defer deleteTestFiles(t, f)
	r, e := cs.GetFileInfo(testPath)
	if e != nil {
		t.Error(e)
	}
	if len(r) != 2 {
		t.Errorf(`Expected 2 files found for GetFileInfo, got %d `, len(r))
	}
	if r[1].ContentType != `application/json` {
		t.Errorf(`Unexpected content type - %s`, r[1].ContentType)
	}
	if r[1].Size != 37 {
		t.Errorf(`Unexpected filesize - %d`, r[1].Size)
	}

}
