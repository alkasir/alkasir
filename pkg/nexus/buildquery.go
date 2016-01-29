package nexus

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/hanjos/nexus/search"
	"github.com/thomasf/lg"
)

// BuildQuery contains all configuration for a building a go cmd
type BuildQuery struct {
	OS      string // target os.
	Arch    string // target arch.
	Cmd     string // target cmd, builds cmd/[cmd]/[cmd].go.
	Version string // version number.
}

func (b *BuildQuery) GetVersions() (Artifacts, error) {
	lg.Infof("Getting versions of %s", b.ArtifactDisplayName())

	artifacts, err := nexusClient.Artifacts(
		search.InRepository{
			RepositoryID: repoID,
			Criteria: search.ByCoordinates{
				GroupID:    "com.alkasir",
				ArtifactID: b.Cmd,
				Classifier: b.Classifier(),
			},
		},
	)

	if err != nil {
		return nil, err
	}

	var result Artifacts
	for _, na := range artifacts {
		a, err := newArtifact(na)
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}

	sort.Sort(result)
	return result, nil
}

// Quickrunner for the latest archived release
func (q *BuildQuery) GetMatchingBuildBinary() (string, error) {
	artifact, err := q.getVersion()
	if err != nil {
		return "", err
	}
	return q.GetBinary(artifact)
}

// ArtifactID constructs a cmd-os-arch string
func (b *BuildQuery) ArtifactDisplayName() string {
	return fmt.Sprintf("%s-%s-%s", b.Cmd, b.OS, b.Arch)
}

// Classifier constructs a os-arch string
func (b *BuildQuery) Classifier() string {
	return fmt.Sprintf("%s-%s", b.OS, b.Arch)
}

// GetBinary downloads artifact, extracts archive and returns the path to the
// extracted executable. If the file already exits it is not downloaded.
func (q *BuildQuery) GetBinary(artifact *Artifact) (string, error) {
	cmdGlob, err := q.cmdGlob()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(artifact.Path()); os.IsNotExist(err) {
		err = artifact.Download()
		if err != nil {
			lg.Fatal(err)
		}
	}
	gp, err := artifact.GlobPath(cmdGlob)
	if err != nil {
		return "", err
	}

	if len(gp) > 0 {
		return gp[0], nil
	}

	err = artifact.Extract()
	if err != nil {
		// retry download if extraction fails
		err = artifact.Download()
		if err != nil {
			return "", err

		}
		err := artifact.Extract()
		if err != nil {
			return "", err
		}
	}

	gp, err = artifact.GlobPath(cmdGlob)
	if err != nil {
		return "", err
	}

	if len(gp) < 1 {
		lg.Fatalf("no glob match for '%s' in %s %s", cmdGlob, artifact.Version, q.ArtifactDisplayName())
	}
	return gp[0], nil
}

func (b *BuildQuery) cmdGlob() (string, error) {

	switch b.Cmd {
	case "alkasir-gui", "alkasir-central", "alkasir-downloader", "alkasir-admin":
		return "/alkasir*", nil
	case "alkasir-client":
		if b.OS == "windows" {
			return "/alkasir*.exe", nil
		}
		return "/alkasir*/alkasir*", nil
	}
	return "", fmt.Errorf("%s is not a supported command", b.Cmd)

}

func (b *BuildQuery) packaging() string {
	if b.OS == "windows" {
		return "zip"
	}
	if b.OS == "darwin" && b.Cmd == "alkasir-gui-osxapp" {
		return "zip"
	}
	return "tar.gz"
}

func (b *BuildQuery) getMasterSnapshot(repoID string) (*Artifact, error) {
	artifacts, err := nexusClient.Artifacts(
		search.InRepository{
			RepositoryID: repoID,
			Criteria: search.ByCoordinates{
				GroupID:    "com.alkasir",
				Version:    "SNAPSHOT",
				ArtifactID: b.Cmd,
				Classifier: b.Classifier(),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if len(artifacts) != 1 {

		lg.Infoln(artifacts)
		return nil, fmt.Errorf("found more than one result: %v", artifacts)
	}
	return &Artifact{Artifact: artifacts[0]}, nil
}

func (b *BuildQuery) getVersion() (*Artifact, error) {
	lg.Infof("Getting versions of %s", b.ArtifactDisplayName())
	artifacts, err := nexusClient.Artifacts(
		search.InRepository{
			RepositoryID: repoID,
			Criteria: search.ByCoordinates{
				GroupID:    "com.alkasir",
				Version:    b.Version,
				ArtifactID: b.Cmd,
				Classifier: b.Classifier(),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if len(artifacts) != 1 {
		lg.Infoln(artifacts)
		return nil, errors.New("one match expected")
	}
	return &Artifact{Artifact: artifacts[0]}, nil
}
