// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gcloud

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/librarian/internal/config"
	"github.com/googleapis/librarian/internal/sources"
	"github.com/googleapis/librarian/internal/testhelper"
)

const testGoogleapisDir = "../../testdata/googleapis"

func TestGenerate(t *testing.T) {
	// sidekickgcloud.Generate calls out to protoc to build a
	// FileDescriptorSet from the protos.
	testhelper.RequireCommand(t, "protoc")

	for _, test := range []struct {
		name    string
		library *config.Library
		golden  string
	}{
		{
			name: "publicca",
			library: &config.Library{
				Name: "publicca",
				APIs: []*config.API{{Path: "google/cloud/security/publicca/v1"}},
			},
			golden: "testdata/publicca",
		},
		{
			name: "parallelstore",
			library: &config.Library{
				Name: "parallelstore",
				APIs: []*config.API{{Path: "google/cloud/parallelstore/v1"}},
				Gcloud: &config.GcloudSurface{
					SupportsStarUpdateMasks: false,
					RootIsHidden:            true,
					HelpText: &config.GcloudHelpTextRules{
						ServiceRules: []*config.GcloudHelpTextRule{
							{
								Selector: "google.cloud.parallelstore.v1.Parallelstore",
								HelpText: &config.GcloudHelpTextElement{
									Brief:       "Manage Parallelstore resources",
									Description: "Manage Parallelstore resources",
								},
							},
						},
						MessageRules: []*config.GcloudHelpTextRule{
							{
								Selector: "google.cloud.parallelstore.v1.Instance",
								HelpText: &config.GcloudHelpTextElement{
									Brief:       "Manage Parallelstore instance resources",
									Description: "Manage Parallelstore instance resources.",
								},
							},
						},
						MethodRules: []*config.GcloudHelpTextRule{
							{
								Selector: "google.cloud.parallelstore.v1.Parallelstore.ListInstances",
								HelpText: &config.GcloudHelpTextElement{
									Brief:       "List Parallelstore instances",
									Description: "List Parallelstore instances.",
									Examples:    []string{"To list all instances in particular location `us-central1-a` run:\n\n$ {command} --location=us-central1-a"},
								},
							},
							{
								Selector: "google.cloud.parallelstore.v1.Parallelstore.GetInstance",
								HelpText: &config.GcloudHelpTextElement{
									Brief:       "Gets details of a single Parallelstore instance",
									Description: "Gets details of a single Parallelstore instance.",
									Examples:    []string{"To get the details of a single instance `my-instance` in location `us-central1-a` run:\n\n$ {command} my-instance --location=us-central1-a"},
								},
							},
							{
								Selector: "google.cloud.parallelstore.v1.Parallelstore.CreateInstance",
								HelpText: &config.GcloudHelpTextElement{
									Brief:       "Creates a Parallelstore instance",
									Description: "Creates a Parallelstore instance.",
									Examples:    []string{"To create an instance `my-instance` in location `us-central1-a` with 12000 Gib capacity run:\n\n$ {command} my-instance --capacity-gib=12000 --location=us-central1-a"},
								},
							},
							{
								Selector: "google.cloud.parallelstore.v1.Parallelstore.UpdateInstance",
								HelpText: &config.GcloudHelpTextElement{
									Brief:       "Updates the parameters of a single Parallelstore instance",
									Description: "Updates the parameters of a single Parallelstore instance.",
									Examples:    []string{"To update the description of an instance `my-instance` in location `us-central1-a` run:\n\n$ {command} my-instance --location=us-central1-a --description=\"<updated description>\""},
								},
							},
							{
								Selector: "google.cloud.parallelstore.v1.Parallelstore.DeleteInstance",
								HelpText: &config.GcloudHelpTextElement{
									Brief:       "Deletes a single Parallelstore instance",
									Description: "Deletes a single Parallelstore instance.",
									Examples:    []string{"To delete an instance `my-instance` run:\n\n$ {command} my-instance"},
								},
							},
							{
								Selector: "google.cloud.parallelstore.v1.Parallelstore.ImportData",
								HelpText: &config.GcloudHelpTextElement{
									Brief:       "Imports data from Cloud Storage to Parallelstore instance.",
									Description: "Imports data from Cloud Storage to Parallelstore instance.",
									Examples:    []string{"To import data from `gs://my-bucket` storage to `my-instance` run:\n\n$ {command} my-instance --location=us-central-a --source-gcs-bucket-uri=gs://my_bucket --destination-parallelstore-path='/'"},
								},
							},
							{
								Selector: "google.cloud.parallelstore.v1.Parallelstore.ExportData",
								HelpText: &config.GcloudHelpTextElement{
									Brief:       "Exports data from Parallelstore instance to Cloud Storage.",
									Description: "Exports data from Parallelstore instance to Cloud Storage.",
									Examples:    []string{"To export data from `my-instance` to `gs://my-bucket` storage  run:\n\n$ {command} my-instance --location=us-central-a --destination-gcs-bucket-uri=gs://my-bucket --source-parallelstore-path='/'"},
								},
							},
						},
					},
					OutputFormatting: []*config.GcloudOutputFormatting{
						{
							Selector: "google.cloud.parallelstore.v1.Parallelstore.ListInstances",
							Format:   "table(name,\n      capacityGib:label=Capacity,\n      description,\n      createTime,\n      updateTime,\n      state,\n      network,\n      reserved_ip_range,\n      accessPoints.join(\",\"))",
						},
					},
					CommandOperationsConfig: []*config.GcloudCommandOperationsConfig{
						{
							Selector:               "google.cloud.parallelstore.v1.Parallelstore.ImportData",
							DisplayOperationResult: true,
						},
						{
							Selector:               "google.cloud.parallelstore.v1.Parallelstore.ExportData",
							DisplayOperationResult: true,
						},
					},
				},
			},
			golden: "testdata/parallelstore",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			out := t.TempDir()
			test.library.Output = out
			srcs := &sources.Sources{Googleapis: testGoogleapisDir}
			if err := Generate(t.Context(), test.library, srcs); err != nil {
				t.Fatal(err)
			}
			compareDirs(t, test.golden, out)
		})
	}
}

func TestGenerate_Errors(t *testing.T) {
	for _, test := range []struct {
		name       string
		googleapis string
		apiPath    string
	}{
		{
			name:       "missing googleapis dir",
			googleapis: "nonexistent_googleapis_dir",
			apiPath:    "google/cloud/security/publicca/v1",
		},
		{
			name:       "missing api path",
			googleapis: testGoogleapisDir,
			apiPath:    "google/cloud/does/not/exist/v1",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			library := &config.Library{
				Name:   "publicca",
				Output: t.TempDir(),
				APIs:   []*config.API{{Path: test.apiPath}},
			}
			srcs := &sources.Sources{Googleapis: test.googleapis}
			if err := Generate(t.Context(), library, srcs); err == nil {
				t.Error("Generate() error = nil, want error")
			}
		})
	}
}

func TestCollectProtos(t *testing.T) {
	apiPath := "google/cloud/security/publicca/v1"
	abs, err := filepath.Abs(testGoogleapisDir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := collectProtos(abs, apiPath)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.ToSlash(filepath.Join(apiPath, "resources.proto")),
		filepath.ToSlash(filepath.Join(apiPath, "service.proto")),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestFindServiceConfig(t *testing.T) {
	apiPath := "google/cloud/security/publicca/v1"
	abs, err := filepath.Abs(testGoogleapisDir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := findServiceConfig(abs, apiPath)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(abs, apiPath, "publicca_v1.yaml")
	if got != want {
		t.Errorf("findServiceConfig() = %q, want %q", got, want)
	}
}

// compareDirs walks goldenDir and gotDir and fails the test on any file
// mismatch, missing file, or extra file.
func compareDirs(t *testing.T, goldenDir, gotDir string) {
	t.Helper()
	goldenFiles := collectFiles(t, goldenDir)
	gotFiles := collectFiles(t, gotDir)

	for rel, goldenPath := range goldenFiles {
		gotPath, ok := gotFiles[rel]
		if !ok {
			t.Errorf("%s: missing in output", rel)
			continue
		}
		want, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(gotPath)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(string(want), string(got)); diff != "" {
			t.Errorf("%s mismatch (-want +got):\n%s", rel, diff)
		}
	}
	for rel := range gotFiles {
		if _, ok := goldenFiles[rel]; !ok {
			t.Errorf("%s: extra file generated", rel)
		}
	}
}

func collectFiles(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = path
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
