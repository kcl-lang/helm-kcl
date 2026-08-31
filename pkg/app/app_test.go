// Copyright 2026 The helm-kcl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"net/url"
	"testing"
)

// TestIsRemoteChartURL exercises the URL-vs-local-path discrimination used by
// App.renderManifests. The previous implementation decided between the two
// loaders based on whether url.Parse returned an error, which is almost
// never the case for any reasonable input string; URLs and local paths
// alike produced a nil parse error and silently fell through to the local
// loader. See https://github.com/kcl-lang/helm-kcl/issues/58.
func TestIsRemoteChartURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"https url", "https://charts.example.com/foo-1.0.0.tgz", true},
		{"http url", "http://charts.example.com/foo-1.0.0.tgz", true},
		{"oci url", "oci://registry.example.com/charts/foo:1.0.0", true},
		// The bug-report URL — a GitHub tree link should still be
		// classified as remote so the remote loader is at least given
		// the chance to handle it (or fail with a meaningful error).
		{"github tree url", "https://github.com/nginxinc/kubernetes-ingress/tree/v3.1.0/deployments/helm-chart", true},

		// Local paths parse successfully but have no scheme; they must
		// fall through to the local loader.
		{"absolute local path", "/etc/charts/foo", false},
		{"relative local path", "./charts/foo", false},
		{"parent-relative local path", "../workload-charts", false},
		{"bare directory name", "workload-charts", false},

		// Schemes that url.Parse accepts but that we do not treat as
		// remote Helm chart URLs.
		{"file url", "file:///tmp/charts/foo.tgz", false},
		{"mailto url", "mailto:dev@example.com", false},

		// Malformed input — url.Parse returns an error, so the local
		// loader is the safe fallback.
		{"malformed url", "://no-scheme", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.in)
			got := err == nil && isRemoteChartURL(u)
			if got != tc.want {
				t.Fatalf("isRemoteChartURL(%q) = %v, want %v (parse err: %v)",
					tc.in, got, tc.want, err)
			}
		})
	}
}
