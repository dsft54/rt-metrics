package main

import (
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func Test_addBuiltinPasses(t *testing.T) {
	tests := []struct {
		name string
		aP   []*analysis.Analyzer
	}{
		{
			name: "normal deep equal",
			aP:   []*analysis.Analyzer{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addBuiltinPasses(tt.aP); len(got) <= 0 {
				t.Error("Builtin passes cannot be picked", got)
			}
		})
	}
}

func Test_addStaticCheck(t *testing.T) {
	tests := []struct {
		name string
		aP   []*analysis.Analyzer
	}{
		{
			name: "normal deep equal",
			aP:   []*analysis.Analyzer{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addStaticCheck(tt.aP)
			Sadded, QFadded, STadded := false, false, false
			for _, v := range got {
				if strings.Contains(v.Name, "S1000") {
					Sadded = true
				}
				if strings.Contains(v.Name, "QF1001") {
					QFadded = true
				}
				if strings.Contains(v.Name, "ST1000") {
					STadded = true
				}
			}
			if !Sadded && !QFadded && !STadded {
				t.Error("Specific staticcheck tests cannot be picked", got)
			}
		})
	}
}

func Test_addCustomCheck(t *testing.T) {
	tests := []struct {
		name string
		aP   []*analysis.Analyzer
	}{
		{
			name: "normal deep equal",
			aP:   []*analysis.Analyzer{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addCustomCheck(tt.aP); len(got) <= 0 {
				t.Error("Builtin passes cannot be picked", got)
			}
		})
	}
}
