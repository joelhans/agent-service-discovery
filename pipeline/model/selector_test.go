package model

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var reSrString = regexp.MustCompile(`^{[^{}]+}$`)

func TestExactSelector_String(t *testing.T) {
	sr := exactSelector("selector")

	assert.True(t, reSrString.MatchString(sr.String()))
}

func TestNegSelector_String(t *testing.T) {
	srs := []Selector{
		exactSelector("selector"),
		negSelector{exactSelector("selector")},
		orSelector{
			lhs: exactSelector("selector"),
			rhs: exactSelector("selector")},
		orSelector{
			lhs: orSelector{lhs: exactSelector("selector"), rhs: negSelector{exactSelector("selector")}},
			rhs: orSelector{lhs: exactSelector("selector"), rhs: negSelector{exactSelector("selector")}},
		},
		andSelector{
			lhs: andSelector{lhs: exactSelector("selector"), rhs: negSelector{exactSelector("selector")}},
			rhs: andSelector{lhs: exactSelector("selector"), rhs: negSelector{exactSelector("selector")}},
		},
	}

	for i, sr := range srs {
		neg := negSelector{sr}
		assert.True(t, reSrString.MatchString(neg.String()), "selector num %d", i+1)
	}
}

func TestOrSelector_String(t *testing.T) {
	sr := orSelector{
		lhs: orSelector{lhs: exactSelector("selector"), rhs: negSelector{exactSelector("selector")}},
		rhs: orSelector{lhs: exactSelector("selector"), rhs: negSelector{exactSelector("selector")}},
	}

	assert.True(t, reSrString.MatchString(sr.String()))
}

func TestAndSelector_String(t *testing.T) {
	sr := andSelector{
		lhs: andSelector{lhs: exactSelector("selector"), rhs: negSelector{exactSelector("selector")}},
		rhs: andSelector{lhs: exactSelector("selector"), rhs: negSelector{exactSelector("selector")}},
	}

	assert.True(t, reSrString.MatchString(sr.String()))
}

func TestExactSelector_Match(t *testing.T) {
	matchTests := struct {
		tags Tags
		srs  []exactSelector
	}{
		tags: Tags{"a": {}, "b": {}},
		srs: []exactSelector{
			"a",
			"b",
		},
	}
	notMatchTests := struct {
		tags Tags
		srs  []exactSelector
	}{
		tags: Tags{"a": {}, "b": {}},
		srs: []exactSelector{
			"c",
			"d",
		},
	}

	for i, sr := range matchTests.srs {
		assert.Truef(t, sr.Matches(matchTests.tags), "match selector num %d", i+1)
	}
	for i, sr := range notMatchTests.srs {
		assert.Falsef(t, sr.Matches(notMatchTests.tags), "not match selector num %d", i+1)
	}
}

func TestNegSelector_Match(t *testing.T) {
	matchTests := struct {
		tags Tags
		srs  []negSelector
	}{
		tags: Tags{"a": {}, "b": {}},
		srs: []negSelector{
			{exactSelector("c")},
			{exactSelector("d")},
		},
	}
	notMatchTests := struct {
		tags Tags
		srs  []negSelector
	}{
		tags: Tags{"a": {}, "b": {}},
		srs: []negSelector{
			{exactSelector("a")},
			{exactSelector("b")},
		},
	}

	for i, sr := range matchTests.srs {
		assert.Truef(t, sr.Matches(matchTests.tags), "match selector num %d", i+1)
	}
	for i, sr := range notMatchTests.srs {
		assert.Falsef(t, sr.Matches(notMatchTests.tags), "not match selector num %d", i+1)
	}
}

func TestOrSelector_Match(t *testing.T) {
	matchTests := struct {
		tags Tags
		srs  []orSelector
	}{
		tags: Tags{"a": {}, "b": {}},
		srs: []orSelector{
			{
				lhs: orSelector{lhs: exactSelector("c"), rhs: exactSelector("d")},
				rhs: orSelector{lhs: exactSelector("e"), rhs: exactSelector("b")},
			},
		},
	}
	notMatchTests := struct {
		tags Tags
		srs  []orSelector
	}{
		tags: Tags{"a": {}, "b": {}},
		srs: []orSelector{
			{
				lhs: orSelector{lhs: exactSelector("c"), rhs: exactSelector("d")},
				rhs: orSelector{lhs: exactSelector("e"), rhs: exactSelector("f")},
			},
		},
	}

	for i, sr := range matchTests.srs {
		assert.Truef(t, sr.Matches(matchTests.tags), "match selector num %d", i+1)
	}
	for i, sr := range notMatchTests.srs {
		assert.Falsef(t, sr.Matches(notMatchTests.tags), "not match selector num %d", i+1)
	}
}

func TestAndSelector_Match(t *testing.T) {
	matchTests := struct {
		tags Tags
		srs  []andSelector
	}{
		tags: Tags{"a": {}, "b": {}, "c": {}, "d": {}},
		srs: []andSelector{
			{
				lhs: andSelector{lhs: exactSelector("a"), rhs: exactSelector("b")},
				rhs: andSelector{lhs: exactSelector("c"), rhs: exactSelector("d")},
			},
		},
	}
	notMatchTests := struct {
		tags Tags
		srs  []andSelector
	}{
		tags: Tags{"a": {}, "b": {}, "c": {}, "d": {}},
		srs: []andSelector{
			{
				lhs: andSelector{lhs: exactSelector("a"), rhs: exactSelector("b")},
				rhs: andSelector{lhs: exactSelector("c"), rhs: exactSelector("z")},
			},
		},
	}

	for i, sr := range matchTests.srs {
		assert.Truef(t, sr.Matches(matchTests.tags), "match selector num %d", i+1)
	}
	for i, sr := range notMatchTests.srs {
		assert.Falsef(t, sr.Matches(notMatchTests.tags), "not match selector num %d", i+1)
	}
}

func TestParseSelector(t *testing.T) {
	tests := map[string]struct {
		wantSelector Selector
		wantErr      bool
	}{
		"":    {wantSelector: trueSelector{}},
		"a":   {wantSelector: exactSelector("a")},
		"Z":   {wantSelector: exactSelector("Z")},
		"a_b": {wantSelector: exactSelector("a_b")},
		"a=b": {wantSelector: exactSelector("a=b")},
		"!a":  {wantSelector: negSelector{exactSelector("a")}},
		"a b": {wantSelector: andSelector{lhs: exactSelector("a"), rhs: exactSelector("b")}},
		"a|b": {wantSelector: orSelector{lhs: exactSelector("a"), rhs: exactSelector("b")}},
		"*":   {wantSelector: trueSelector{}},
		"!*":  {wantSelector: negSelector{trueSelector{}}},
		"a b !c d|e f": {
			wantSelector: andSelector{
				lhs: andSelector{
					lhs: andSelector{
						lhs: andSelector{lhs: exactSelector("a"), rhs: exactSelector("b")},
						rhs: negSelector{exactSelector("c")},
					},
					rhs: orSelector{
						lhs: exactSelector("d"),
						rhs: exactSelector("e"),
					},
				},
				rhs: exactSelector("f"),
			},
		},
		"!":      {wantErr: true},
		"a !":    {wantErr: true},
		"a!b":    {wantErr: true},
		"0a":     {wantErr: true},
		"a b c*": {wantErr: true},
		"__":     {wantErr: true},
		"a|b|c*": {wantErr: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			sr, err := ParseSelector(name)

			if test.wantErr {
				assert.Nil(t, sr)
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.wantSelector, sr)
			}
		})
	}
}

func TestMustParseSelector(t *testing.T) {
	tests := []string{
		"!",
		"a !",
		"a!b",
		"0a",
		"a b c*",
		"__",
		"a|b|c*",
	}

	for _, test := range tests {
		f := func() {
			MustParseSelector(test)
		}
		assert.Panicsf(t, f, test)
	}
}
