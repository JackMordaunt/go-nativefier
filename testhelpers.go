package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

const (
	onePixelPng = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="
)

type mockInferrer struct {
	Error bool
}

func (i mockInferrer) Infer(url string, prefs []string) (*Icon, error) {
	var (
		icon *Icon
		err  error
	)
	if !i.Error {
		pngData := []byte(onePixelPng)
		dst := make([]byte, len(pngData))
		_, err := base64.StdEncoding.Decode(dst, pngData)
		if err != nil {
			panic(fmt.Sprintf("decoding test png: %v", err))
		}
		data := bytes.NewBuffer(dst)
		icon = &Icon{
			Source: "https://url/to/icon.png",
			Data:   data,
			Size:   data.Len(),
			Mime:   "image/png",
			Ext:    "png",
		}
	} else {
		err = fmt.Errorf("failed to infer icon")
	}
	return icon, err
}

// the caching is overkill, nevertheless, it should be fast if the
// same pattern is tested for many times or the list of log messages is large.
type testLogger struct {
	Msgs  []string
	cache map[string]bool
}

func (logger *testLogger) Logf(format string, values ...interface{}) {
	logger.Msgs = append(logger.Msgs, fmt.Sprintf(format, values...))
}

func (logger *testLogger) Contains(pattern string) (contains bool) {
	if res, ok := logger.checkCache(pattern); ok {
		return res
	}
	defer logger.setCache(pattern, contains)
	for _, msg := range logger.Msgs {
		if strings.Contains(msg, pattern) {
			contains = true
		}
	}
	return contains
}

func (logger *testLogger) String() string {
	buf := bytes.NewBufferString("[")
	for ii, msg := range logger.Msgs {
		buf.WriteString("\t")
		buf.WriteString(msg)
		if ii < len(logger.Msgs)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

func (logger *testLogger) checkCache(pattern string) (contains bool, ok bool) {
	if logger.cache == nil {
		return false, false
	}
	if v, ok := logger.cache[pattern]; ok {
		return v, ok
	}
	return false, false
}

func (logger *testLogger) setCache(pattern string, contains bool) {
	if logger.cache == nil {
		logger.cache = map[string]bool{}
	}
	logger.cache[pattern] = contains
}
