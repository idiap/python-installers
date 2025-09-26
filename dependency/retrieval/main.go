// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

// Part of this code is taken from their respective buildpacks

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/joshuatcasey/libdependency/buildpack_config"
	"github.com/joshuatcasey/libdependency/retrieve"
	"github.com/joshuatcasey/libdependency/upstream"
	"github.com/joshuatcasey/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/fs"
)

type PyPiProductMetadataRaw struct {
	Releases map[string][]struct {
		PackageType string            `json:"packagetype"`
		URL         string            `json:"url"`
		UploadTime  string            `json:"upload_time_iso_8601"`
		Digests     map[string]string `json:"digests"`
	} `json:"releases"`
}

type PyPiRelease struct {
	version      *semver.Version
	SourceURL    string
	UploadTime   time.Time
	SourceSHA256 string
}

func (release PyPiRelease) Version() *semver.Version {
	return release.version
}

func getAllVersionsForInstaller(installer string) retrieve.GetAllVersionsFunc {
	return func() (versionology.VersionFetcherArray, error) {

		var pypiMetadata PyPiProductMetadataRaw
		err := upstream.GetAndUnmarshal(fmt.Sprintf("https://pypi.org/pypi/%s/json", installer), &pypiMetadata)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve new versions from upstream: %w", err)
		}

		var allVersions versionology.VersionFetcherArray

		for version, releasesForVersion := range pypiMetadata.Releases {
			for _, release := range releasesForVersion {
				if release.PackageType != "sdist" {
					continue
				}

				fmt.Printf("Parsing semver version %s\n", version)

				newVersion, err := semver.NewVersion(version)
				if err != nil {
					continue
				}

				uploadTime, err := time.Parse(time.RFC3339, release.UploadTime)
				if err != nil {
					return nil, fmt.Errorf("could not parse upload time '%s' as date for version %s: %w", release.UploadTime, version, err)
				}

				allVersions = append(allVersions, PyPiRelease{
					version:      newVersion,
					SourceSHA256: release.Digests["sha256"],
					SourceURL:    release.URL,
					UploadTime:   uploadTime,
				})
			}
		}

		return allVersions, nil
	}
}

func generatePipMetadata(versionFetcher versionology.VersionFetcher) ([]versionology.Dependency, error) {
	version := versionFetcher.Version().String()
	pipRelease, ok := versionFetcher.(PyPiRelease)
	if !ok {
		return nil, errors.New("expected a PyPiRelease")
	}

	configMetadataDependency := cargo.ConfigMetadataDependency{
		CPE:            fmt.Sprintf("cpe:2.3:a:pypa:pip:%s:*:*:*:*:python:*:*", version),
		ID:             "pip",
		Licenses:       retrieve.LookupLicenses(pipRelease.SourceURL, upstream.DefaultDecompress),
		PURL:           retrieve.GeneratePURL("pip", version, pipRelease.SourceSHA256, pipRelease.SourceURL),
		Source:         pipRelease.SourceURL,
		SourceChecksum: fmt.Sprintf("sha256:%s", pipRelease.SourceSHA256),
		Stacks:         []string{"*"},
		Version:        version,
	}

	return versionology.NewDependencyArray(configMetadataDependency, "noarch")
}

func generatePipenvMetadata(versionFetcher versionology.VersionFetcher) ([]versionology.Dependency, error) {
	version := versionFetcher.Version().String()
	pipenvRelease, ok := versionFetcher.(PyPiRelease)
	if !ok {
		return nil, errors.New("expected a PyPiRelease")
	}

	configMetadataDependency := cargo.ConfigMetadataDependency{
		CPE:            fmt.Sprintf("cpe:2.3:a:python-pipenv:pipenv:%s:*:*:*:*:python:*:*", version),
		Checksum:       fmt.Sprintf("sha256:%s", pipenvRelease.SourceSHA256),
		ID:             "pipenv",
		Licenses:       retrieve.LookupLicenses(pipenvRelease.SourceURL, upstream.DefaultDecompress),
		Name:           "Pipenv",
		PURL:           retrieve.GeneratePURL("pipenv", version, pipenvRelease.SourceSHA256, pipenvRelease.SourceURL),
		Source:         pipenvRelease.SourceURL,
		SourceChecksum: fmt.Sprintf("sha256:%s", pipenvRelease.SourceSHA256),
		Stacks:         []string{"*"},
		URI:            pipenvRelease.SourceURL,
		Version:        version,
	}

	return []versionology.Dependency{{
		ConfigMetadataDependency: configMetadataDependency,
		SemverVersion:            versionFetcher.Version(),
	}}, nil
}

func generatePoetryMetadata(versionFetcher versionology.VersionFetcher) ([]versionology.Dependency, error) {
	version := versionFetcher.Version().String()
	poetryRelease, ok := versionFetcher.(PyPiRelease)
	if !ok {
		return nil, errors.New("expected a PyPiRelease")
	}

	configMetadataDependency := cargo.ConfigMetadataDependency{
		CPE:            fmt.Sprintf("cpe:2.3:a:python-poetry:poetry:%s:*:*:*:*:python:*:*", version),
		Checksum:       fmt.Sprintf("sha256:%s", poetryRelease.SourceSHA256),
		ID:             "poetry",
		Licenses:       retrieve.LookupLicenses(poetryRelease.SourceURL, upstream.DefaultDecompress),
		Name:           "Poetry",
		PURL:           retrieve.GeneratePURL("poetry", version, poetryRelease.SourceSHA256, poetryRelease.SourceURL),
		Source:         poetryRelease.SourceURL,
		SourceChecksum: fmt.Sprintf("sha256:%s", poetryRelease.SourceSHA256),
		Stacks:         []string{"*"},
		URI:            poetryRelease.SourceURL,
		Version:        version,
	}

	return []versionology.Dependency{{
		ConfigMetadataDependency: configMetadataDependency,
		SemverVersion:            versionFetcher.Version(),
	}}, nil
}

// Taken from libdependency.retrieve.retrieval
// https://github.com/joshuatcasey/libdependency/blob/main/retrieve/retrieval.go
func toWorkflowJson(item any) (string, error) {
	if bytes, err := json.Marshal(item); err != nil {
		return "", err
	} else {
		return string(bytes), nil
	}
}

func main() {
	buildpackTomlPathUsage := "full path to the buildpack.toml file, using only one of camelCase, snake_case, or dash_case"
	var buildpackTomlPath string
	var output string

	flag.StringVar(&buildpackTomlPath, "buildpack_toml_path", buildpackTomlPath, buildpackTomlPathUsage)
	flag.StringVar(&output, "output", "", "filename for the output JSON metadata")
	flag.Parse()

	exists, err := fs.Exists(buildpackTomlPath)

	if err != nil {
		panic(err)
	} else if !exists {
		panic(fmt.Errorf("could not locate buildpack.toml at '%s'", buildpackTomlPath))
	}

	if output == "" {
		panic("output is required")
	}

	config, err := buildpack_config.ParseBuildpackToml(buildpackTomlPath)
	if err != nil {
		panic(err)
	}

	metadataGeneratorMap := map[string]retrieve.GenerateMetadataFunc{
		"pip":    generatePipMetadata,
		"pipenv": generatePipenvMetadata,
		"poetry": generatePoetryMetadata,
	}

	var dependencies []versionology.Dependency

	for id, generateMetadata := range metadataGeneratorMap {
		newVersions, err := retrieve.GetNewVersionsForId(id, config, getAllVersionsForInstaller(id))
		if err != nil {
			panic(err)
		}

		// This loop is taken from GenerateAllMetadata
		// https://github.com/joshuatcasey/libdependency/blob/main/retrieve/retrieval.go
		for _, version := range newVersions {
			metadata, err := generateMetadata(version)
			if err != nil {
				panic(err)
			}

			var targets []string
			for _, metadatum := range metadata {
				targets = append(targets, metadatum.Target)
			}
			fmt.Printf("Generating metadata for %s, with targets [%s]\n", version.Version().String(), strings.Join(targets, ", "))
			dependencies = append(dependencies, metadata...)
		}
	}

	metadataJson, err := toWorkflowJson(dependencies)
	if err != nil {
		panic(fmt.Errorf("unable to marshall metadata json, with error=%w", err))
	}

	if err = os.WriteFile(output, []byte(metadataJson), os.ModePerm); err != nil {
		panic(fmt.Errorf("cannot write to %s: %w", output, err))
	} else {
		fmt.Printf("Wrote metadata to %s\n", output)
	}
}
