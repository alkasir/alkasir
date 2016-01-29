package client

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/thomasf/lg"
	"github.com/alkasir/alkasir/pkg/service"
	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/alkasir/alkasir/pkg/upgradebin"
)

var (
	artifactNameMu sync.Mutex
	artifactName   string
)

func SetUpgradeArtifact(name string) {
	artifactNameMu.Lock()
	artifactName = fmt.Sprintf("%s-%s-%s", name, runtime.GOOS, runtime.GOARCH)
	artifactNameMu.Unlock()
}

// StartBinaryUpgradeChecker checks for binary upgrades when the connection is
// up and on a schedule.
//
// This function runs in it's own goroutine.
func StartBinaryUpgradeChecker() {
	connectionEventListener := make(chan service.ConnectionHistory)
	uChecker, _ := NewUpdateChecker("binary")
	service.AddListener(connectionEventListener)
	for {
		select {
		// Update when the transport connection comes up
		case event := <-connectionEventListener:
			if event.IsUp() {
				uChecker.Activate()
				uChecker.UpdateNow()
			}

		// Update by request of the update checker
		case request := <-uChecker.RequestC:
			err := upgradeBinaryCheck()
			// err := fakeUpgradeBinaryCheck()
			if err != nil {
				lg.Errorln(err)
				request.ResponseC <- UpdateError
			} else {
				request.ResponseC <- UpdateSuccess
			}
		}
	}
}

func upgradeBinaryCheck() error {
	if !upgradeEnabled {
		lg.Infoln("binary upgrades are disabled using the command line flag")
		return nil
	}
	artifactNameMu.Lock()
	artifact := artifactName
	artifactNameMu.Unlock()

	cl, err := NewRestClient()
	if err != nil {
		return err
	}
	// TODO: check for current artifact + version (need to add artifact id to cmd's)
	res, found, err := cl.CheckBinaryUpgrade(shared.BinaryUpgradeRequest{
		Artifact:    artifact,
		FromVersion: VERSION,
	})
	if err != nil {
		return err
	}
	if !found {
		lg.Infoln("no update found")
		return nil
	}
	lg.Warningf("found update %+v", res)

	httpclient, err := service.NewTransportHTTPClient()
	if err != nil {
		return err
	}

	opts, err := upgradebin.NewUpdaterOptions(res, shared.UpgradeVerificationPublicKey)
	if err != nil {
		return err
	}
	

	URL := fmt.Sprintf("https://central.server.domain/u/%s/%s/%s", artifact, VERSION, res.Version)
	lg.Infoln("downloading %s", URL)
	resp, err := httpclient.Get(URL)
	if err != nil {
		lg.Errorln(err)
		return err
	}

	defer resp.Body.Close()
	err = upgradebin.Apply(resp.Body, opts)
	if err != nil {
		lg.Errorln(err)
		// will be retried the next time the client starts
		return nil
	}

	return nil

}
