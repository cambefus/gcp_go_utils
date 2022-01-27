package storage

/*
	wrapper for GCP cloud storage
	requires initialization with a secrets structure to obtain configuration values
*/
import (
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
)

type CStore struct {
	client      *storage.Client
	bucket      *storage.BucketHandle
	credentials []byte
}

var defaultClient *storage.Client

// NewCStore - creates an object suitable for accessing a predefined Bucket in GCP Cloud Storage
func NewCStore(cred []byte, bucketName string) (*CStore, error) {
	this := new(CStore)
	if len(cred) == 0 || len(bucketName) == 0 {
		return nil, errors.New(`invalid parameter(s)`)
	}
	var err error
	this.client, err = storage.NewClient(context.Background(), option.WithCredentialsJSON(cred))
	if err == nil && this.client != nil {
		if defaultClient == nil {
			defaultClient = this.client
		}
		this.bucket = this.client.Bucket(bucketName)
	}
	this.credentials = cred
	return this, err
}

// NewCStoreP - creates an object suitable for accessing a predefined Bucket in GCP Cloud Storage, assumes permissions exist, no credentials required
// note - it will reuse the default client for any subsequent calls.
func NewCStoreP(bucketName string) (*CStore, error) {
	this := new(CStore)
	if len(bucketName) == 0 {
		return nil, errors.New(`invalid parameter(s)`)
	}
	if defaultClient == nil {
		var err error
		defaultClient, err = storage.NewClient(context.Background())
		if err != nil {
			return nil, err
		}
	}
	this.client = defaultClient
	this.bucket = this.client.Bucket(bucketName)
	return this, nil
}

// GetFiles - return list of files within specified bucket / path
func (cs *CStore) GetFiles(path string) ([]string, error) {
	var result []string
	var q = storage.Query{Prefix: path}
	it := cs.bucket.Objects(context.Background(), &q)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		result = append(result, objAttrs.Name)
	}
	return result, nil
}

// GetFilteredFiles - walk through the objects within specified path, passing each filename to a function
// continues as long as pf result is true
func (cs *CStore) GetFilteredFiles(path string, pf func(oa *storage.ObjectAttrs) bool) error {
	var q = storage.Query{Prefix: path}
	it := cs.bucket.Objects(context.Background(), &q)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if !pf(objAttrs) {
			break
		}
	}
	return nil
}

// DeleteOldFiles - delete all files within the specified path that are older than age in hours
// returns # files deleted. Will stop if an error occurs
func (cs *CStore) DeleteOldFiles(path string, ageh int) (int, error) {
	timeLimit := time.Now().Add(time.Duration(ageh*-1) * time.Hour)
	result := 0
	oa, err := cs.GetFileInfo(path)
	if err != nil {
		return 0, err
	}
	for _, attrs := range oa {
		if attrs.Created.Before(timeLimit) {
			e2 := cs.DeleteCloudFile(attrs.Name)
			if e2 == nil {
				result++
			} else {
				return result, e2
			}
		}
	}
	return result, nil
}

// GetFileInfo - returns slice of files within the bucket
func (cs *CStore) GetFileInfo(path string) ([]storage.ObjectAttrs, error) {
	var result []storage.ObjectAttrs
	var q = storage.Query{Prefix: path}
	it := cs.bucket.Objects(context.Background(), &q)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		result = append(result, *objAttrs)
	}
	return result, nil
}

// GetFilesWithSuffix - returns list of files from the the specified path where the file name ends with suffix
func (cs *CStore) GetFilesWithSuffix(path string, suffix string) ([]string, error) {
	var result []string
	var q = storage.Query{Prefix: path}
	it := cs.bucket.Objects(context.Background(), &q)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(objAttrs.Name, suffix) {
			result = append(result, objAttrs.Name)
		}
	}
	return result, nil
}

// GetFileReader - remember to close the Reader after use. returns error if file not found
// second parameter is file size in bytes, if found
func (cs *CStore) GetFileReader(fn string) (*storage.Reader, int64, error) {
	it := cs.bucket.Object(fn)
	var fsize int64
	ita, e1 := it.Attrs(context.Background())
	if e1 != nil {
		return nil, fsize, e1
	} else {
		fsize = ita.Size
	}
	r, err := it.NewReader(context.Background())
	if err != nil {
		return nil, fsize, err
	}
	return r, fsize, nil
}

// DeleteCloudFile -
func (cs *CStore) DeleteCloudFile(fn string) error {
	it := cs.bucket.Object(fn)
	err := it.Delete(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// FileExists -
func (cs *CStore) FileExists(fn string) bool {
	it := cs.bucket.Object(fn)
	_, err := it.Attrs(context.Background())
	return err == nil
}

// WriteFile creates a text file in Google Cloud Storage.
func (cs *CStore) WriteFile(fn string, content string) error {
	return cs.WriteCloudFile(fn, []byte(content), "text/plain")
}

// WriteCloudFile - write data to a file in the google cloud
// fn is the dest filename/path
// content - what to write
// ftype is the Mime contentType
func (cs *CStore) WriteCloudFile(fn string, content []byte, ftype string) error {
	wc := cs.bucket.Object(fn).NewWriter(context.Background())
	wc.ContentType = ftype
	if _, err := wc.Write(content); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return nil
}

// CopyFile - from/to locations within the cloud
func (cs *CStore) CopyFile(srcName string, destcs *CStore, dest string) error {
	s := cs.bucket.Object(srcName)
	d := destcs.bucket.Object(dest)

	_, err := d.CopierFrom(s).Run(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// CreateDownloadURL - create a signed, time limited url to access the specified file
// https://cloud.google.com/storage/docs/access-control/signing-urls-manually
// https://cloud.google.com/storage/docs/authentication/canonical-requests
func (cs *CStore) CreateDownloadURL(minutes int, path string) (string, error) {
	f := cs.bucket.Object(path)
	conf, err := google.JWTConfigFromJSON(cs.credentials)
	if err != nil {
		return "", err
	}
	opts := &storage.SignedURLOptions{
		Method:         "GET",
		GoogleAccessID: conf.Email,
		PrivateKey:     conf.PrivateKey,
		Expires:        time.Now().Add(time.Duration(minutes) * time.Minute),
	}
	u, err := storage.SignedURL(f.BucketName(), path, opts)
	if err != nil {
		return "", err
	}
	return u, nil
}

// DownloadFiles - assumes list of files contains folder names
// dest is local file path
func (cs *CStore) DownloadFiles(files []string, dest string) error {
	for _, fn := range files {
		cr, _, err := cs.GetFileReader(fn)
		if err != nil {
			return err
		}
		if cr != nil {
			c2, e1 := ioutil.ReadAll(cr)
			if e1 != nil {
				return e1
			}
			fb := filepath.Base(fn)
			e2 := ioutil.WriteFile(dest+fb, c2, 0644)
			if e2 != nil {
				return e2
			}
		}
	}
	return nil
}
