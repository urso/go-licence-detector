// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package detector

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.elastic.co/go-licence-detector/dependency"
)

func TestDetect(t *testing.T) {
	testCases := []struct {
		name             string
		includeIndirect  bool
		overrides        dependency.Overrides
		wantDependencies func() *dependency.List
		wantErr          bool
	}{
		{
			name:            "All",
			includeIndirect: true,
			overrides: map[string]dependency.Info{
				"github.com/gorhill/cronexpr": {Name: "github.com/gorhill/cronexpr", LicenceType: "GPL-3.0"},
			},
			wantDependencies: func() *dependency.List {
				return &dependency.List{
					Indirect: mkIndirectDeps(),
					Direct:   mkDirectDeps(),
				}
			},
		},
		{
			name:            "DirectOnly",
			includeIndirect: false,
			overrides: map[string]dependency.Info{
				"github.com/gorhill/cronexpr": {Name: "github.com/gorhill/cronexpr", LicenceType: "GPL-3.0"},
			},
			wantDependencies: func() *dependency.List {
				return &dependency.List{
					Direct: mkDirectDeps(),
				}
			},
		},
		{
			name:            "WithOverrides",
			includeIndirect: true,
			overrides: map[string]dependency.Info{
				"github.com/davecgh/go-spew":         {Name: "github.com/davecgh/go-spew", URL: "http://example.com/go-spew"},
				"github.com/russross/blackfriday/v2": {Name: "github.com/russross/blackfriday/v2", LicenceType: "MIT"},
				"github.com/gorhill/cronexpr":        {Name: "github.com/gorhill/cronexpr", LicenceType: "GPL-3.0"},
			},
			wantDependencies: func() *dependency.List {
				deps := &dependency.List{}

				for _, d := range mkIndirectDeps() {
					d := d
					if d.Name == "github.com/davecgh/go-spew" {
						d.URL = "http://example.com/go-spew"
					}
					deps.Indirect = append(deps.Indirect, d)
				}

				for _, d := range mkDirectDeps() {
					d := d
					if d.Name == "github.com/russross/blackfriday/v2" {
						d.LicenceType = "MIT"
					}
					deps.Direct = append(deps.Direct, d)
				}

				return deps
			},
		},
		{
			name:            "WithValidLicenceFileOverride",
			includeIndirect: true,
			overrides: map[string]dependency.Info{
				"github.com/gorhill/cronexpr": {Name: "github.com/gorhill/cronexpr", LicenceFile: "GPLv3"},
			},
			wantDependencies: func() *dependency.List {
				return &dependency.List{
					Indirect: mkIndirectDeps(),
					Direct:   mkDirectOverridenDeps(),
				}
			},
		},
		{
			name:            "WithInvalidLicenceFileOverride",
			includeIndirect: true,
			overrides: map[string]dependency.Info{
				"github.com/davecgh/go-spew":         {Name: "github.com/davecgh/go-spew", LicenceFile: "/path/to/nowhere"},
				"github.com/russross/blackfriday/v2": {Name: "github.com/russross/blackfriday/v2", LicenceFile: "/path/to/nowhere"},
			},
			wantErr: true,
		},

		{
			name:            "LicenceNotAllowed",
			includeIndirect: true,
			overrides: map[string]dependency.Info{
				"github.com/davecgh/go-spew":         {Name: "github.com/davecgh/go-spew", LicenceType: "Totally Legit License 2.0"},
				"github.com/russross/blackfriday/v2": {Name: "github.com/russross/blackfriday/v2", LicenceType: "MIT"},
				"github.com/davecgh/go-gk":           {Name: "github.com/davecgh/go-spew", LicenceType: "UNKNOWN"},
			},
			wantErr: true,
		},
	}

	// create classifier
	classifier, err := NewClassifier("")
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.Open("testdata/deps.json")
			require.NoError(t, err)
			defer f.Close()

			rules, err := LoadRules("testdata/rules.json")
			require.NoError(t, err)

			gotDependencies, err := Detect(f, classifier, rules, tc.overrides, tc.includeIndirect)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantDependencies(), gotDependencies)
		})
	}
}

func mkIndirectDeps() []dependency.Info {
	return []dependency.Info{
		{
			Name:        "github.com/davecgh/go-spew",
			Version:     "v1.1.0",
			VersionTime: "2016-10-29T20:57:26Z",
			Dir:         "testdata/github.com/davecgh/go-spew@v1.1.0",
			LicenceType: "ISC",
			LicenceFile: "testdata/github.com/davecgh/go-spew@v1.1.0/LICENCE.txt",
			URL:         "https://github.com/davecgh/go-spew",
		},
		{
			Name:        "github.com/dgryski/go-minhash",
			Version:     "v0.0.0-20170608043002-7fe510aff544",
			VersionTime: "2017-06-08T04:30:02Z",
			Dir:         "testdata/github.com/dgryski/go-minhash@v0.0.0-20170608043002-7fe510aff544",
			LicenceType: "MIT",
			LicenceFile: "testdata/github.com/dgryski/go-minhash@v0.0.0-20170608043002-7fe510aff544/licence",
			URL:         "https://github.com/dgryski/go-minhash",
		},
		{
			Name:        "github.com/dgryski/go-spooky",
			Version:     "v0.0.0-20170606183049-ed3d087f40e2",
			VersionTime: "2017-06-06T18:30:49Z",
			Dir:         "testdata/github.com/dgryski/go-spooky@v0.0.0-20170606183049-ed3d087f40e2",
			LicenceType: "MIT",
			LicenceFile: "testdata/github.com/dgryski/go-spooky@v0.0.0-20170606183049-ed3d087f40e2/COPYING",
			URL:         "https://github.com/dgryski/go-spooky",
		},
	}
}

func mkDirectDeps() []dependency.Info {
	return []dependency.Info{
		{
			Name:        "github.com/ekzhu/minhash-lsh",
			Version:     "v0.0.0-20171225071031-5c06ee8586a1",
			VersionTime: "2017-12-25T07:10:31Z",
			Dir:         "testdata/github.com/ekzhu/minhash-lsh@v0.0.0-20171225071031-5c06ee8586a1",
			LicenceType: "MIT",
			LicenceFile: "testdata/github.com/ekzhu/minhash-lsh@v0.0.0-20171225071031-5c06ee8586a1/licence.txt",
			URL:         "https://github.com/ekzhu/minhash-lsh",
		},
		{
			Name:        "github.com/russross/blackfriday/v2",
			Version:     "v2.0.1",
			VersionTime: "2018-09-20T17:16:15Z",
			Dir:         "testdata/github.com/russross/blackfriday/v2@v2.0.1",
			LicenceType: "BSD-2-Clause",
			LicenceFile: "testdata/github.com/russross/blackfriday/v2@v2.0.1/LICENSE.rst",
			URL:         "https://github.com/russross/blackfriday",
		},
		{
			Name:        "github.com/gorhill/cronexpr",
			Version:     "v0.0.0-20161205141322-d520615e531a",
			VersionTime: "2016-12-05T14:13:22Z",
			Dir:         "testdata/github.com/gorhill/cronexpr@v0.0.0-20161205141322-d520615e531a",
			LicenceType: "GPL-3.0",
			LicenceFile: "",
			URL:         "https://github.com/gorhill/cronexpr",
		},
	}
}

func mkDirectOverridenDeps() []dependency.Info {
	return []dependency.Info{
		{
			Name:        "github.com/ekzhu/minhash-lsh",
			Version:     "v0.0.0-20171225071031-5c06ee8586a1",
			VersionTime: "2017-12-25T07:10:31Z",
			Dir:         "testdata/github.com/ekzhu/minhash-lsh@v0.0.0-20171225071031-5c06ee8586a1",
			LicenceType: "MIT",
			LicenceFile: "testdata/github.com/ekzhu/minhash-lsh@v0.0.0-20171225071031-5c06ee8586a1/licence.txt",
			URL:         "https://github.com/ekzhu/minhash-lsh",
		},
		{
			Name:        "github.com/russross/blackfriday/v2",
			Version:     "v2.0.1",
			VersionTime: "2018-09-20T17:16:15Z",
			Dir:         "testdata/github.com/russross/blackfriday/v2@v2.0.1",
			LicenceType: "BSD-2-Clause",
			LicenceFile: "testdata/github.com/russross/blackfriday/v2@v2.0.1/LICENSE.rst",
			URL:         "https://github.com/russross/blackfriday",
		},
		{
			Name:        "github.com/gorhill/cronexpr",
			Version:     "v0.0.0-20161205141322-d520615e531a",
			VersionTime: "2016-12-05T14:13:22Z",
			Dir:         "testdata/github.com/gorhill/cronexpr@v0.0.0-20161205141322-d520615e531a",
			LicenceType: "GPL-3.0",
			LicenceFile: "testdata/github.com/gorhill/cronexpr@v0.0.0-20161205141322-d520615e531a/GPLv3",
			URL:         "https://github.com/gorhill/cronexpr",
		},
	}
}

func TestDetermineURL(t *testing.T) {
	testCases := []struct {
		name     string
		override string
		modPath  string
		want     string
	}{
		{
			name:     "WithOverride",
			override: "https://go.elast.co/dep",
			modPath:  "github.com/elastic/dep/path",
			want:     "https://go.elast.co/dep",
		},
		{
			name:    "WithNonGitHubPath",
			modPath: "go.uber.org/zap",
			want:    "https://go.uber.org/zap",
		},
		{
			name:    "WithValidGitHubPath",
			modPath: "github.com/elastic/cloud-on-k8s",
			want:    "https://github.com/elastic/cloud-on-k8s",
		},
		{
			name:    "WithInvalidGitHubPath",
			modPath: "github.com/elastic/cloud-on-k8s/api/v1/elasticsearch",
			want:    "https://github.com/elastic/cloud-on-k8s",
		},
		{
			name:    "WithK8sPath",
			modPath: "k8s.io/apimachinery",
			want:    "https://github.com/kubernetes/apimachinery",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			have := determineURL(tc.override, tc.modPath)
			require.Equal(t, tc.want, have)
		})
	}
}
