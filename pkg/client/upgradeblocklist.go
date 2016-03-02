package client

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/alkasir/alkasir/pkg/client/internal/config"
	"github.com/alkasir/alkasir/pkg/client/ui"
	"github.com/alkasir/alkasir/pkg/pac"
	"github.com/alkasir/alkasir/pkg/service"
	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/nu7hatch/gouuid"
	"github.com/thomasf/lg"
)

// StartBlocklistUpgrader react to certain conitions for when the list of
// blocked urls should be updated.
//
// This function runs in it's own goroutine.
func StartBlocklistUpgrader() {
	connectionEventListener := make(chan service.ConnectionHistory)
	uChecker, _ := NewUpdateChecker("blocklist")
	service.AddListener(connectionEventListener)
	currentCountry := clientconfig.Get().Settings.Local.CountryCode
	checkCountrySettingC := time.NewTicker(2 * time.Second)
	defer checkCountrySettingC.Stop()
loop:
	for {
		select {
		// Update when the transport connection comes up
		case event := <-connectionEventListener:
			if event.IsUp() {
				uChecker.Activate()
				uChecker.UpdateNow()
			}

		// Tell updatechecker to request update when user changes country settings
		case <-checkCountrySettingC.C:
			conf := clientconfig.Get()
			if currentCountry != conf.Settings.Local.CountryCode {
				currentCountry = conf.Settings.Local.CountryCode
				uChecker.UpdateNow()
			}

		// Update by request of the update checker
		case request := <-uChecker.RequestC:
			conf := clientconfig.Get()
			if conf.Settings.Local.CountryCode == "__" {
				lg.V(9).Infoln("Country is __, skipping blocklist updates")
				continue loop
			}
			currentCountry = conf.Settings.Local.CountryCode
			n, err := upgradeBlockList()
			if err != nil {
				lg.Errorf("blocklist update cc:%s err:%v", currentCountry, err)
				ui.Notify("blocklist_update_error_message")
				request.ResponseC <- UpdateError
			} else {
				lg.V(5).Infof("blocklist update success cc:%s, got %d entries", currentCountry, n)
				ui.Notify("blocklist_update_success_message")
				request.ResponseC <- UpdateSuccess
			}
		}
	}
}

func upgradeBlockList() (int, error) {
	conf := clientconfig.Get()
	restclient, err := NewRestClient()
	if err != nil {
		return 0, err
	}
	// a new update ID is sent every week
	var updateID string
	savedID := conf.Settings.LastID
	_, week := time.Now().ISOWeek()
	nowID := (week % 3) + 1
	if nowID != savedID {
		id, err := uuid.NewV4()
		if err != nil {
			lg.Errorln("could not generate client id", err)
		} else {
			if savedID == 0 {
				updateID = (fmt.Sprintf("%d:%s", savedID, id.String()))
			} else {
				updateID = id.String()
			}
		}
	}
	if updateID != "" {
		lg.Infoln("sending UpdateID", updateID)
	}
	req := shared.UpdateHostlistRequest{
		ClientAddr:    getPublicIPAddr(),
		UpdateID:      updateID,
		ClientVersion: VERSION,
	}
	resp, err := restclient.UpdateHostlist(req)
	n := len(resp.Hosts)
	if err != nil {
		return n, err
	}
	if nowID != savedID {
		err := clientconfig.Update(func(conf *clientconfig.Config) error {
			conf.Settings.LastID = nowID
			return nil
		})

		if err != nil {
			lg.Errorln(err)
		}
		err = clientconfig.Write()
		if err != nil {
			return n, err
		}

	}

	newHosts := resp.Hosts
	prevHosts := conf.BlockedHostsCentral.Hosts

	sort.Strings(newHosts)
	sort.Strings(prevHosts)

	if !reflect.DeepEqual(newHosts, prevHosts) {
		err := clientconfig.Update(func(conf *clientconfig.Config) error {
			lg.V(2).Infoln("hosts list updated and changed")
			lg.V(20).Infoln("hosts received: %v", newHosts)
			conf.BlockedHostsCentral.Hosts = newHosts
			pac.UpdateBlockedList(conf.BlockedHostsCentral.Hosts, conf.BlockedHosts.Hosts)
			lastBlocklistChange = time.Now()
			return nil
		})
		if err != nil {
			lg.Errorln(err)
		}
	} else {
		lg.V(19).Infoln("Hostlists equal after update")
	}
	return n, nil
}
