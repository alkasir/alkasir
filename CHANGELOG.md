# 0.4.8 - unreleased

# 0.4.7 - (2016-09-21) 

- MacOS Sierra compatibility
- Added Alkasir MacOS Dock icons
- Fixed some potential client lockup errors
- Updated to go 1.7.1
- Extracted systray package to use vendored version
- Upgraded vendored libraries

# 0.4.6 (2016-05-31)

- Fix http timeouts not being used, this resolves a lot of situation where the
  client can lock up or not reconnect to transports.
- Fix bug where reported site samples for measurements which had not finished
  when the user clicks the "send" button.

# 0.4.5

- Fixed bug where browser pac doesnt update after block list is updated

# 0.4.4

- Add server side measurements analysis pipeline [central]
- Fix transport/connection deadlocks [client]
- Fix all js/jsx linting errors [browser]
- Fix exported debug log encryption for binary downloads [dist]
- Compile using Go 1.6 [build]

# 0.4.3

- Add timeouts in shutdown procedure so that quit if there are other problems [client]

# 0.4.2

- Display version number on the settings page [client]
- Debug export logs are now encrypted [client]
- Add english and farsi in app docs/help pages [client/browser extension] 
- Misc improvements and fixes to docs system [client/browser extension] 
- Chrome extension exporter not allowed to run in multiple instances concurrently [client] 
- Add suggestion session token to  [central]
- Record http status code and http redirects in http header measurement [client/central]

# 0.4.1/0.4.0

- Changelog starts here

