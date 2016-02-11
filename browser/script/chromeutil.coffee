settings = require './settings'

getAuthKey = () ->
  localStorage.authKey

gotoPage = (url) ->
  fulurl = chrome.extension.getURL(url)
  chrome.tabs.getAllInWindow undefined, (tabs) ->
    for i of tabs
      if !tabs.hasOwnProperty(i)
        continue
      tab = tabs[i]
      if tab.url == fulurl
        chrome.tabs.update tab.id, selected: true
        return
    chrome.tabs.getSelected null, (tab) ->
      chrome.tabs.create
        url: url
        index: tab.index + 1
      return
    return
  return

setAuthCookie = () ->
  chrome.cookies.set({
    url: "#{settings.baseURL}"
    name: "authKey"
    value: getAuthKey()
    domain: "localhost"
  })

gotoClientPage = (page) ->
  setAuthCookie()
  gotoPage "#{settings.baseURL}/##{page}"


module.exports = {
  enabled: false  # set to true when chrome is enabled
  gotoPage: gotoPage
  gotoClientPage: gotoClientPage

}
