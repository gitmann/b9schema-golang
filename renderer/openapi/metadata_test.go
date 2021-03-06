package openapi

import (
	"github.com/gitmann/b9schema-golang/common/util"
	"strings"
	"testing"
)

// TestNewMetaData validates that metadata is formatted into valid YAML.
func TestNewMetaData(t *testing.T) {
	testCases := []struct {
		name     string
		meta     *MetaData
		wantYAML []string
	}{
		{
			name: "default",
			meta: NewMetaData("", ""),
			wantYAML: []string{
				`openapi: 3.0.0`,
				`info:`,
				`  title: default title`,
				`  version: default version`,
			},
		},
		{
			name: "full",
			meta: &MetaData{
				OpenAPI: OPENAPI_VERSION,
				Info: &InfoObject{
					Title:          "This is the title.",
					Version:        "v1.2.3",
					Description:    "This is a description.",
					TermsOfService: "https://test.tos.site.com/terms",
					Contact: &ContactObject{
						Name:  "Support Team",
						URL:   "https://support.site.com/",
						Email: "support@site.com",
					},
					License: &LicenseObject{
						Name: "This is the license.",
						URL:  "https://license.site.com/",
					},
				},
				Servers: []*ServerObject{
					{
						URL:         "https://www.site.com",
						Description: "Production server.",
					},
					{
						URL:         "https://www.dev.site.com",
						Description: "Development server.",
					},
				},
				ExternalDocs: &ExternalDocumentationObject{
					URL:         "https://test.doc.site.com/path/to/docs",
					Description: "This is the test doc site.",
				},
			},
			wantYAML: []string{
				`openapi: 3.0.0`,
				`info:`,
				`  title: This is the title.`,
				`  version: v1.2.3`,
				`  description: This is a description.`,
				`  termsOfService: https://test.tos.site.com/terms`,
				`  contact:`,
				`    email: support@site.com`,
				`    name: Support Team`,
				`    url: https://support.site.com/`,
				`  license:`,
				`    name: This is the license.`,
				`    url: https://license.site.com/`,
				`externalDocs:`,
				`  description: This is the test doc site.`,
				`  url: https://test.doc.site.com/path/to/docs`,
				`servers:`,
				`  - description: Production server.`,
				`    url: https://www.site.com`,
				`  - description: Development server.`,
				`    url: https://www.dev.site.com`,
			},
		},
	}

	for _, test := range testCases {
		if b, err := test.meta.MarshalYAML("  "); err != nil {
			t.Errorf("TEST_FAIL %s: yaml err=%s", test.name, err)
		} else {
			gotYAML := strings.Split(string(b), "\n")
			test.wantYAML = append(test.wantYAML, ``)

			util.CompareStrings(t, test.name, gotYAML, test.wantYAML)
		}
	}
}
