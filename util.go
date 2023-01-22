// Copyright 2023 The firestore Authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"fmt"
	"strings"

	json "github.com/goccy/go-json"
)

func colorFormat(attr int) json.ColorFormat {
	return json.ColorFormat{
		Header: fmt.Sprintf("\x1b[%dm", attr),
		Footer: "\x1b[0m",
	}
}

var colorScheme = &json.ColorScheme{
	Int:       colorFormat(93),
	Uint:      colorFormat(93),
	Float:     colorFormat(93),
	Bool:      colorFormat(95),
	String:    colorFormat(32),
	Binary:    colorFormat(91),
	ObjectKey: colorFormat(96),
	Null:      colorFormat(37),
}

// CutLast similar strings.Cut, but for the last index of sep.
func CutLast(s, sep string) (before, after string, ok bool) {
	if i := strings.LastIndex(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}

// CutSuffix returns s without the provided ending suffix string
// and reports whether it found the suffix.
// If s doesn't end with suffix, CutSuffix returns s, false.
// If suffix is the empty string, CutSuffix returns s, true.
func CutSuffix(s, suffix string) (before string, found bool) {
	if !strings.HasSuffix(s, suffix) {
		return s, false
	}
	return s[:len(s)-len(suffix)], true
}
