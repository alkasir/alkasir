package nexus

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/hanjos/nexus"
	version "github.com/hashicorp/go-version"
	"github.com/pivotal-golang/archiver/extractor"
	"github.com/thomasf/lg"
)

// Artifact combines nexus artifact with semver features
type Artifact struct {
	*nexus.Artifact
	v *version.Version
}

func newArtifact(a *nexus.Artifact) (*Artifact, error) {
	v, err := version.NewVersion(a.Version)
	if err != nil {
		return nil, err

	}
	return &Artifact{
		Artifact: a,
		v:        v,
	}, nil
}

func (a *Artifact) Dir(ns string) string {
	if ns == "" {
		panic("namespace required")
	}
	return filepath.Join(baseDir, ns, a.GroupID, a.ArtifactID, a.Version, a.Classifier+"-"+a.Extension)
}

// Path returns the path to the downloaded artifact
func (a *Artifact) Path() string {
	return filepath.Join(a.Dir("dl"), path.Base(a.Info().URL))
}

func (a *Artifact) Info() *nexus.ArtifactInfo {

	info, err := nexusClient.InfoOf(a.Artifact)
	if err != nil {
		panic(err)
	}
	return info
}

// Download downloads the artifact to a local directory
func (a *Artifact) Download() error {
	shortURL := a.Info().URL
	shortURL = shortURL[strings.LastIndex(shortURL, "/")+1:]
	lg.Infof("Downloading %s", shortURL)
	err := os.MkdirAll(a.Dir("dl"), 0775)
	if err != nil {
		return err
	}
	err = download(a.Path(), a.Info().URL)
	if err != nil {
		lg.Warningf("Error downloading %s", shortURL)
	} else {
		lg.Infof("Downloaded %s", shortURL)
	}
	return err
}

func (a *Artifact) Extract() error {
	extractor := extractor.NewDetectable()
	dir := a.Dir("extracted")
	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}
	os.MkdirAll(dir, 0775)
	apath, err := filepath.Abs(a.Path())
	if err != nil {
		return err
	}
	lg.V(5).Infof("extracting %s into %s", apath, dir)
	return extractor.Extract(apath, dir)
}

func (a *Artifact) GlobPath(cmdglob string) ([]string, error) {
	baseDir := a.Dir("extracted")
	files, err := filepath.Glob(fmt.Sprintf("%s%s", baseDir, cmdglob))
	if err != nil {
		return nil, err
	}
	return files, nil
}

// Binaries returns a list of patjhs to binaries
func (a *Artifact) Binaries() []string {
	dir := a.Dir("extracted")
	fileList := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if isExecutable(f) {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return fileList
}

// LinkBinaries links any binaries from an extracted archive path to PWD.
func (a *Artifact) LinkBinaries() error {
	bins := a.Binaries()
	for _, file := range bins {
		basename := filepath.Base(file)
		lg.Infoln(file, basename)
		os.Remove(basename)
		os.Symlink(file, basename)
	}
	return nil
}

func (a *Artifact) Run(cmdglob string) error {
	files, err := a.GlobPath(cmdglob)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		lg.Infoln("wrong number of results for glob", files)
		os.Exit(1)
	}
	cmd := exec.Command(files[0], os.Args[2:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// GreaterThan compares the semver version number
func (a *Artifact) GreaterThan(b *Artifact) bool {
	return a.v.GreaterThan(b.v)
}

// LessThan compares the semver version number
func (a *Artifact) LessThan(b *Artifact) bool {
	return a.v.LessThan(b.v)
}

// Equal compares the semver version number, nothing else.
func (a *Artifact) Equal(b *Artifact) bool {
	return a.v.Equal(b.v)
}
