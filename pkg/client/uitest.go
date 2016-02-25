package client

import (
	"fmt"
	"strings"

	"github.com/alkasir/alkasir/pkg/client/internal/config"
	"github.com/alkasir/alkasir/pkg/i18n"
	"github.com/ant0ine/go-json-rest/rest"
)

const UITEST = false

// GetNotifications writes all current notifications.
func GetNotificationsUITest(w rest.ResponseWriter, r *rest.Request) {

	language := clientconfig.Get().Settings.Local.Language

	Has, err := i18n.Hasfunc(language)
	if err != nil {
		panic(err)
	}

	var notifications []Notification

	for _, v := range []string{"blocklist_update_error",
		"blocklist_update_init",
		"blocklist_update_success",
		"blocklist_update_warning",
		"central_connected",
		"central_error",
		"central_warning",
		"extension_client_not_available",
		"extension_connected",
		"extension_disconnected",
		"listen_error",
		"network_connected",
		"network_error",
		"network_timeout_error",
		"server_error",
		"status_error",
		"status_ok",
		"status_warning",
		"suggest_denied",
		"suggest_this",
		"transport_connected",
		"transport_disconnected",
		"transport_error",
		"transport_init",
		"transport_warning",
		"transported_connection_opened",
		"unknown_error",
	} {

		level := "info"
		if strings.Contains(v, "error") {
			level = "danger"
		}

		if strings.Contains(v, "warning") {
			level = "warning"
		}

		if strings.Contains(v, "_ok") || strings.Contains(v, "success") {
			level = "success"
		}

		mkey := fmt.Sprintf("%s_message", v)
		tkey := fmt.Sprintf("%s_title", v)

		if !Has(tkey) && Has(strings.Replace(v, "_ok", "", 1)) {
			tkey = strings.Replace(v, "_ok", "", 1)
		}

		if !Has(tkey) && Has(strings.Replace(v, "_error", "", 1)) {
			tkey = strings.Replace(v, "_error", "", 1)
		}

		if !Has(tkey) && Has(strings.Replace(v, "_warning", "", 1)) {
			tkey = strings.Replace(v, "_warning", "", 1)
		}

		notifications = append(notifications,
			Notification{
				Level:   level,
				Title:   tkey,
				Message: mkey,
				Actions: []NotificationAction{
					{"action_continue", "/suggestions/"},
					{"action_ok", "/setup-guide/"},
					{"action_help", "/setup-guide/"},
				},
			})
	}
	notifications = prepareNotifications(notifications)
	w.WriteJson(notifications)
}
