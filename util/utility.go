package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"time"
)

// UniqueInts returns a unique subset of the int slice provided.
func UniqueInts(input []int) []int {
	u := make([]int, 0, len(input))
	m := make(map[int]bool)
	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}
	return u
}

// IntSliceToCSV - convert a slice of integers into a CSV string
func IntSliceToCSV(ns []int) string {
	if len(ns) == 0 {
		return ""
	}
	// Appr. 3 chars per num plus the comma.
	estimate := len(ns) * 3
	b := make([]byte, 0, estimate)
	for x, n := range ns {
		if x > 0 {
			b = append(b, ',')
		}
		b = strconv.AppendInt(b, int64(n), 10)
	}
	return string(b)
}

// Filter - accept slice of string, return slice with only those entries that return true to passed function
func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func JSONMarshalNoEscape(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

// FileExists - takes a filename, returns bool
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

type CustomLogFormat struct{}

func (lf *CustomLogFormat) Format(entry *log.Entry) ([]byte, error) {
	if entry.Level.String() == "info" {
		return []byte(fmt.Sprintf("%s \n", entry.Message)), nil
	}
	return []byte(fmt.Sprintf("%s: %s \n", entry.Level, entry.Message)), nil
}

func EncodeInteger(key int) string {
	secs := time.Now().Unix()
	r1 := secs % 16
	r2 := (secs / 100) % 16

	mkey := int64(key) + ((r1 + 1) * 300) + r2

	return strings.ToUpper(strconv.FormatInt(r1, 16) + strconv.FormatInt(mkey, 16) + strconv.FormatInt(r2, 16))
}

func DecodeInteger(key string) int {
	if len(key) < 3 {
		return 0
	}
	kl := len(key)
	r1, _ := strconv.ParseInt(key[0:1], 16, 64)
	r2, _ := strconv.ParseInt(key[kl-1:kl], 16, 64)
	k, _ := strconv.ParseInt(key[1:kl-1], 16, 64)
	return int(k - (((r1 + 1) * 300) + r2))
}

// StringsDiff - set difference, result is A - B
func StringsDiff(a, b []string) (diff []string) {
	m := make(map[string]bool, len(b))
	for _, item := range b {
		m[item] = true
	}
	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}
