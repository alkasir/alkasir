package makepatch

import (
	"bytes"
	"crypto"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/agl/ed25519"
	"github.com/inconshreveable/go-update"
	"github.com/kr/binarydist"
	"github.com/thomasf/lg"
	"github.com/alkasir/alkasir/pkg/nexus"
	"github.com/alkasir/alkasir/pkg/upgradebin"
)

// patchJob .
type CreatePatchJob struct {
	Artifact   string
	OldBinary  string
	NewBinary  string
	OldVersion string
	NewVersion string
	PrivateKey string
	PublicKey  string
}

// Patch .
type CreatePatchResult struct {
	job              CreatePatchJob
	Artifact         string `json:"artifact"`
	OldVersion       string `json:"oldVersion"`
	NewVersion       string `json:"newVersion"`
	SHA256Sum        string `json:"sha256sum"`
	ED25519Signature string `json:"ed25519sig"`
	DiffFile         string `json:"-"`
}

var PatchHistoryAmountMax = 50

const (
	diffsDir = "diffs"
)

func RunPatchesCreate(queries []nexus.BuildQuery, privateKey string, publicKey string, nWorkers int) ([]CreatePatchResult, error) {
	if nWorkers < 1 {
		nWorkers = 1
	}
	jobC := make(chan CreatePatchJob, 6)

	var creators sync.WaitGroup
	for _, v := range queries {
		creators.Add(1)
		go func(b nexus.BuildQuery) {
			defer creators.Done()
			err := createJobs(b, jobC, privateKey, publicKey)
			if err != nil {
				lg.Fatal(err)
			}
		}(v)

	}
	resC := make(chan CreatePatchResult, 0)
	var differs sync.WaitGroup
	for workerN := 0; workerN < nWorkers; workerN++ {
		differs.Add(1)
		go func() {
			defer differs.Done()
			for job := range jobC {
				res, err := CreatePatch(job)
				if err != nil {
					lg.Fatal(err)
				}
				resC <- res
			}
		}()
	}

	go func() {
		creators.Wait()
		close(jobC)
		differs.Wait()
		close(resC)
	}()

	var patches []CreatePatchResult
	defer func() {
		for _, p := range patches {
			os.Remove(p.job.NewBinary)
			os.Remove(p.job.OldBinary)

		}
	}()
	for pr := range resC {
		patches = append(patches, pr)

	}

	return patches, nil
}

func createJobs(q nexus.BuildQuery, jobC chan CreatePatchJob, privateKey string, publicKey string) error {
	versions, err := q.GetVersions()
	if err != nil {
		lg.Fatal(err)
	}

	sort.Sort(versions)
	sort.Sort(sort.Reverse(versions))
	if len(versions) < 2 {
		return errors.New("too few versions")
	}
	latestVersion := versions[0]
	lg.V(20).Infoln("latest version", latestVersion)
	{
		if len(versions) > PatchHistoryAmountMax+1 {
			versions = versions[1 : PatchHistoryAmountMax+1]
		} else {
			versions = versions[1:]
		}
	}
	lg.V(20).Infoln("old versions",
		versions)

	// TODO: reimplement this check so that upgrade processing can be resumed.
	// if _, err := os.Stat(jsonname); err == nil {
	// 	lg.Infof("%s exists, skipping processing", jsonname)
	// 	return nil
	// }

	latestBinPath, err := q.GetBinary(latestVersion)
	if err != nil {
		return err
	}

	lg.Infof("creating patchJobs for %s %s %s",
		latestVersion.ArtifactID,
		latestVersion.Classifier,
		latestVersion.Version,
	)
	for _, v := range versions {
		bp, err := q.GetBinary(v)
		if err != nil {
			return err
		}

		j := CreatePatchJob{
			Artifact:   fmt.Sprintf("%s-%s", latestVersion.ArtifactID, latestVersion.Classifier),
			OldBinary:  bp,
			NewBinary:  latestBinPath,
			NewVersion: latestVersion.Version,
			OldVersion: v.Version,
			PrivateKey: privateKey,
			PublicKey:  publicKey,
		}
		lg.V(10).Infof("sending created job %s", j.Artifact)
		jobC <- j
		lg.V(10).Infof("sent job %s", j.Artifact)
	}
	lg.Infof("all jobs created for %s", q.ArtifactDisplayName())
	return nil
}

func CreatePatch(job CreatePatchJob) (CreatePatchResult, error) {
	var emptyResult = CreatePatchResult{}

	logstr := fmt.Sprintf("%s -> %s (%s)", job.OldVersion, job.NewVersion, job.Artifact)

	outfile := filepath.Join(diffsDir,
		job.Artifact,
		job.OldVersion,
		job.NewVersion)

	lg.V(10).Infoln("load new binary into memory", logstr)
	var newData []byte
	{
		var err error
		newData, err = ioutil.ReadFile(job.NewBinary)
		if err != nil {
			return emptyResult, nil
		}
	}

	lg.V(10).Infoln("shasum new", logstr)
	var latestSumBytes []byte
	{
		h := sha256.New()
		h.Write(newData)
		latestSumBytes = h.Sum(nil)
	}

	latestSum := base64.RawURLEncoding.EncodeToString(
		latestSumBytes)

	lg.V(10).Infoln("sign new", logstr)
	var latestSig string
	{
		privateKey, err := upgradebin.DecodePrivateKey([]byte(job.PrivateKey))
		if err != nil {
			return emptyResult, err
		}
		latestSig = base64.RawURLEncoding.EncodeToString(
			ed25519.Sign(privateKey, latestSumBytes)[:])

	}

	var diff []byte
	{
		lg.V(10).Infoln("generate diff", logstr)
		var patch bytes.Buffer
		newFile := bytes.NewReader(newData)
		oldFile, err := os.Open(job.OldBinary)
		if err != nil {
			return emptyResult, nil
		}
		defer oldFile.Close()
		if err := binarydist.Diff(oldFile, newFile, &patch); err != nil {
			return emptyResult, nil
		}
		diff = patch.Bytes()
	}

	{
		lg.V(10).Infoln("write diff", logstr)
		if err := os.MkdirAll(filepath.Dir(outfile), 0775); err != nil {
			return emptyResult, err
		}
		if err := ioutil.WriteFile(outfile, diff, 0775); err != nil {
			return emptyResult, err
		}
	}

	res := CreatePatchResult{
		job:              job,
		Artifact:         job.Artifact,
		NewVersion:       job.NewVersion,
		OldVersion:       job.OldVersion,
		SHA256Sum:        latestSum,
		ED25519Signature: latestSig,
		DiffFile:         outfile,
	}

	data, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return emptyResult, err
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s.json", outfile), data, 0775)
	if err != nil {
		return emptyResult, err
	}

	lg.Infoln("done", logstr)
	return res, nil
}

func testPatch(pr CreatePatchResult, publicKey string) error {
	lg.Infof("verifying %s   %s>%s", pr.Artifact, pr.OldVersion, pr.NewVersion)
	tmpfile := fmt.Sprintf("/tmp/%s-%s-o", pr.Artifact, pr.OldVersion)
	err := cp(tmpfile, pr.job.OldBinary)
	if err != nil {
		lg.Fatal(err)
	}

	defer func() {
		err = os.Remove(tmpfile)
		if err != nil {
			lg.Errorln(err)
		}
	}()

	sum, err := base64.RawURLEncoding.DecodeString(pr.SHA256Sum)
	if err != nil {
		return err
	}

	sig, err := upgradebin.DecodeSignature(pr.ED25519Signature)
	if err != nil {
		return err
	}
	pub, err := upgradebin.DecodePublicKey([]byte(publicKey))
	if err != nil {
		return err
	}

	opts := update.Options{
		Patcher:    update.NewBSDiffPatcher(),
		Verifier:   upgradebin.NewED25519Verifier(),
		Hash:       crypto.SHA256,
		Checksum:   sum,
		Signature:  sig[:],
		PublicKey:  pub,
		TargetPath: tmpfile,
	}

	diffFile, err := os.Open(pr.DiffFile)
	if err != nil {
		return err
	}
	defer diffFile.Close()

	err = update.Apply(diffFile, opts)
	if err != nil {
		return err
	}

	return nil
}

// copy file (does not copy attributes)
func cp(dst, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}
