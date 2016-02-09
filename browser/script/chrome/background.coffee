### global chrome, localStorage, _, signals ###

_ = require 'lodash'
signals = require 'signals'

Actions = require '../Actions.coffee'
Stores = require '../Stores.coffee'
ImmutableDate = require "bloody-immutable-date"
chromeutil = require '../chromeutil.coffee'
settings = require '../settings.coffee'

settings.baseURL = "http://localhost:8899"

alkasirDebug = false

resURL = (resource) ->
  return settings.baseURL + "/api" + resource


appState =
  clientConnectionOk: false
  transportOk: false
  blocklistUpdateOk: false
  lastBlocklistChange: 'init'
  lastSuggestionOpened: new ImmutableDate()
  traffic: {throughput: 0}


appSignals = 
  stateUpdated: new signals.Signal
  clientConnected: new signals.Signal
  clientDisconnected: new signals.Signal
  clientPingOk: new signals.Signal
  requestError: new signals.Signal
  maybeBlocked: new signals.Signal
  suggestionCreated: new signals.Signal



parseDomain = (url) ->
  try
    match = url.match(/^[\w-]+:\/*\[?([\w\.:-]+)\]?(?::\d+)?/)
    if match.length < 2
      return url
  catch err
    return url
  match[1]


# merges appState and dispatches state update events
setState = (obj) ->
  prevState = _.clone(appState)
  newState = _.defaults(obj, appState)
  # TODO: for now we always send the new state regardless of content changes
  # if (_.isEqual(newState, prevState)) {
  #    console.log("state is same");
  #     return
  # }
  appState = _.clone(newState)
  appSignals.stateUpdated.dispatch _.clone(newState), _.clone(prevState)
  return


if localStorage.lastSuggestionOpened?
  setState {
    lastSuggestionOpened: new ImmutableDate(localStorage.lastSuggestionOpened)}


iconSet = (str) ->
  icon = path: 'images/on.png'
  if str == 'off'
    icon.path = 'images/off.png'
  if str == 'transported'
    icon.path = 'images/transported.png'
  if str == 'warning'
    icon.path = 'images/warning.png'
  if currentIcon == icon.path
    return
  chrome.browserAction.setIcon icon
  currentIcon = icon.path
  return

getAuthKey = () ->
  localStorage.authKey

addAuthKey = (req) ->
  if localStorage.authKey
    req.setRequestHeader("Authorization", "Bearer " + localStorage.authKey)
  req


# Check if client is running
clientRunningCheck = ->
  req = new XMLHttpRequest
  url = resURL '/version/'
  req.open 'GET', url, true
  addAuthKey(req)
  req.onreadystatechange = (->
    if @readyState == 4
      if @status == 200
        setState clientConnectionOk: true
        appSignals.clientPingOk.dispatch()
      else
        setState clientConnectionOk: false
    return
  ).bind(req)
  req.send null
  return



# Check the client status
clientStatusCheck = ->
  req = new XMLHttpRequest
  url = resURL '/status/summary/'
  req.open 'GET', url, true
  addAuthKey(req)
  req.onreadystatechange = (->
    if @readyState == 4
      if @status == 200
        resp = JSON.parse(@response)
        setState
          transportOk: resp.transportOk
          lastBlocklistChange: resp.lastBlocklistChange
      else
        setState transportOk: false
    return
  ).bind(req)
  req.send null
  return



getTransportTraffic = ->
  req = new XMLHttpRequest
  url = resURL '/transports/traffic/'
  req.open 'GET', url, true
  addAuthKey(req)
  req.onreadystatechange = (->
    if @readyState == 4
      if @status == 200
        resp = JSON.parse(@response)
        setState {traffic: resp}
      else
        appSignals.requestError.dispatch 'Could not fetch traffic status'
    return
  ).bind(req)
  req.send null
  return


pacActive = false

###
# set direct (no) proxy
#
###
directProxy = ->
  config = {mode: 'direct'}
  chrome.proxy.settings.set {
    value: config
    scope: 'regular'
  }, ->
    pacActive = false
  return



###
# set pac script proxy
#
###

pacProxy = ->
  if !getAuthKey()
    return

  pacURL = 'http://localhost:8899/api/pac/' +
    encodeURIComponent(Date.now()) + '/' +
    "?authKey=" + getAuthKey()

  config =
    mode: 'pac_script'
    pacScript:
      mandatory: true
      url: pacURL

  chrome.proxy.settings.set {
    value: config
    scope: 'regular'
  }, () ->
    pacActive = true
  return


chrome.proxy.onProxyError.addListener (details) ->
  console.log "proxy error", details

appSignals.stateUpdated.add (cur, prev) ->
  if cur.clientConnectionOk and !prev.clientConnectionOk
    appSignals.clientConnected.dispatch()
  return



# Dispatch clientDisconnected signal
appSignals.stateUpdated.add (cur, prev) ->
  if !cur.clientConnectionOk and prev.clientConnectionOk
    appSignals.clientDisconnected.dispatch()
  return



if alkasirDebug
  appSignals.stateUpdated.add (currentState, prevState) ->
    console.log '> CUR:', currentState, '> PREV:', prevState
    return



# TODO
chrome.runtime.onInstalled.addListener (details) ->
  if details.reason == 'install'
    localStorage.clientURL = 'http://localhost:8899'
    chromeutil.gotoPage 'options.html'
  return



chrome.commands.onCommand.addListener (command) ->
  if command == 'open-option'
    chromeutil.gotoPage 'options.html'
  return



###
# set the icon on or off
#
###
currentIcon = ''
# Update popup icon according to connection state
appSignals.stateUpdated.add (cur, prev) ->
  if (cur.clientConnectionOk == prev.clientConnectionOk) and
     (cur.traffic.throughput == prev.traffic.throughput) and
     (cur.transportOk == prev.transportOk)
    return 0
  icn = 'off'
  if cur.clientConnectionOk
    if cur.transportOk
      if (cur.traffic.throughput > 1024)
        icn = "transported"
      else
        icn = "on"
    else
      icn = 'warning'
  iconSet icn
  0

clientConnectionNotificationId = null


getNotificationId = (prefix = "notify") ->
  id = Math.floor(Math.random() * 9007199254740992) + 1
  #Stores latest notification ID so that event handlers can access
  #notification when background page is closed.
  # chrome.storage.local.set 'id': id
  return prefix + "-" + id.toString()



# Handle connected to client
appSignals.stateUpdated.add (cur, prev) ->
  if cur.clientConnectionOk and !prev.clientConnectionOk
    if clientConnectionNotificationId != null
      chrome.notifications.clear clientConnectionNotificationId, (->)
    clientConnectionNotificationId = null
    opt =
      type: 'basic'
      title: chrome.i18n.getMessage("extension_connected_title")
      message: chrome.i18n.getMessage("extension_connected_message")
      iconUrl: chrome.extension.getURL 'images/on.png'
      eventTime: Date.now()
    if clientConnectionNotificationId == null
      clientConnectionNotificationId = getNotificationId("cliUp")
      chrome.notifications.create clientConnectionNotificationId, opt, (i) ->
    else
      chrome.notifications.update clientConnectionNotificationId, opt, (i) ->
    return
  return



# Handle (re)connected to transport
appSignals.stateUpdated.add (cur, prev) ->
  # If client state changes, that will be shown instead
  if cur.clientConnectionOk != prev.clientConnectionOk
    return
  # If the client is not connected that will be shown
  if !cur.clientConnectionOk
    return
  if !cur.transportOk
    return
  if cur.transportOk and !prev.transportOk
    if clientConnectionNotificationId != null
      chrome.notifications.clear clientConnectionNotificationId, ->
    clientConnectionNotificationId = null
    # opt =
    #   type: 'basic'
    #   title: chrome.i18n.getMessage("transport_connected_title")
    #   message: chrome.i18n.getMessage("transport_connected_message")
    #   iconUrl: chrome.extension.getURL 'images/on.png'
    #   eventTime: Date.now()
    # if clientConnectionNotificationId == null
    #   clientConnectionNotificationId = getNotificationId()
    #   chrome.notifications.create clientConnectionNotificationId, opt, ->
    # else
    #   chrome.notifications.update clientConnectionNotificationId, opt, ->
  return



# Handle client connection error
appSignals.stateUpdated.add (cur, prev) ->
  if !cur.clientConnectionOk
    if prev.clientConnectionOk
      if clientConnectionNotificationId != null
        chrome.notifications.clear clientConnectionNotificationId, ->
        clientConnectionNotificationId = null
    opt =
      type: 'basic'
      title: chrome.i18n.getMessage("extension_client_not_available_title")
      message: chrome.i18n.getMessage("extension_client_not_available_message")
      priority: 2
      iconUrl: chrome.extension.getURL 'images/off.png'
      eventTime: Date.now()
      buttons: [
        { title: 'Help' }
      ]
    if clientConnectionNotificationId == null
      clientConnectionNotificationId = getNotificationId("cliDown")
      chrome.notifications.create clientConnectionNotificationId, opt, ->
    else
      chrome.notifications.update clientConnectionNotificationId, opt, ->
  return



# Handle transport disconnection erroprs
appSignals.stateUpdated.add (cur, prev) ->
  if cur.clientConnectionOk != prev.clientConnectionOk
    return
  if !cur.clientConnectionOk
    return
  if cur.transportOk == prev.transportOk
    return
  if !cur.transportOk
    if prev.transportOk
      if clientConnectionNotificationId != null
        chrome.notifications.clear clientConnectionNotificationId, ->
        clientConnectionNotificationId = null
    opt =
      type: 'basic'
      title: chrome.i18n.getMessage("transport_warning_title")
      message: chrome.i18n.getMessage("transport_warning_message")
      iconUrl: chrome.extension.getURL 'images/off.png'
      eventTime: Date.now()
      buttons: [
        { title: 'Help' }
      ]
    if clientConnectionNotificationId == null
      clientConnectionNotificationId = getNotificationId("TrDown")
      chrome.notifications.create clientConnectionNotificationId, opt, ->
        if alkasirDebug
          console.log "created transport warning notification", opt
    else
      chrome.notifications.update clientConnectionNotificationId, opt, ->
        if alkasirDebug
          console.log "updatedtransport warning notification", opt
  return


# Handle opening new Suggestion sessions in browser tabs
Stores.suggestions.reset.add (items) ->
  if not items? or items == "null"
    return
  dtNewest = appState.lastSuggestionOpened
  for item in items
    dt = new ImmutableDate(item.CreatedAt)
    if dt.setMilliseconds(0) > appState.lastSuggestionOpened.setMilliseconds(0)
      console.log "bb", dt, appState.lastSuggestionOpened
      appSignals.suggestionCreated.dispatch item
      if dt > dtNewest
        console.log "newer"
        dtNewest = dt

  localStorage.lastSuggestionOpened = dtNewest
  setState {lastSuggestionOpened: dtNewest}


appSignals.suggestionCreated.add (item) ->
  console.log "> getNotificationId >> #{item.ID}"
  chromeutil.gotoClientPage "/suggestions/#{item.ID}/"
  true


appSignals.clientPingOk.add clientStatusCheck
appSignals.clientPingOk.add getTransportTraffic
appSignals.clientPingOk.add Actions.getAllSuggestions


# disable proxy config at start up to get it into a known state.
directProxy()

# Set direct method when client is unreachable
appSignals.clientDisconnected.add directProxy

appSignals.stateUpdated.add (cur, prev) ->
  if !cur.clientConnectionOk || !cur.transportOk
    directProxy()
    return 0

  if !pacActive and cur.clientConnectionOk and cur.transportOk
    pacProxy()
    return 0

  if cur.lastBlocklistChange != prev.lastBlocklistChange
    pacProxy()
    return 0

  return 0

# TODO: It seems to be impossible form isidide an extension see if
# request has been sent through a proxy.. These events will still
#  be useful for blocked page detection.
#
# chrome.webRequest.onBeforeSendHeaders.addListener(
#     function(details) {
#         "use strict";
#             if (!_.startsWith(details.url, "http://localhost:8899")) {
#                 // console.log(details);
#             }
#         },
#         {urls: ["<all_urls>"]},
#     ["requestHeaders"]);



# # This is only a check for any http status > 400
# chrome.webRequest.onCompleted.addListener ((details) ->
#   if !appState.clientConnectionOk
#     return
#   if !_.startsWith(details.url, 'http://localhost:8899')
#     if details.statusCode >= 400
#       appSignals.maybeBlocked.dispatch details
#   # if (details.ip === "127.0.0.1" ) {
#   # }
#   return
# ), { urls: [ '<all_urls>' ] }, [ 'responseHeaders' ]



# # Handle showing notification when a new connection is opened
# to a blocked host.
# maybeAdds = {}
# appSignals.maybeBlocked.add (value) ->
#   console.log "maybeblocked is disabled", value
#   return true
#   hostname = parseDomain(value.url)
#   id = 'maybeBlocked: ' + hostname
#   opt =
#     type: 'basic'
#     title: hostname + ' might be blocked'
#     message: value.url + ' might be blocked'
#     # note: removed icon
#     iconUrl: chrome.extension.getURL 'images/maybeblocked.png'
#     eventTime: Date.now()
#     priority: 2
#     buttons: [
#       { title: chrome.i18n.getMessage("action_continue") }
#       { title: chrome.i18n.getMessage("action_ignore") + " " + hostname }
#     ]
#   chrome.notifications.create id, opt, ->
#   maybeAdds[id] = value
#   return



# chrome.notifications
# .onButtonClicked
# .addListener (notificationId, buttonIndex) ->
#   req = new XMLHttpRequest
#   url = apiurl + '/suggestions/'
#   req.open 'POST', url, true
#   req.setRequestHeader 'Content-Type', 'application/json;charset=UTF-8'
#   req.onreadystatechange = (->
#     if @readyState == 4
#       if @status == 200
#         obj = JSON.parse(@responseText)
#         chrome.tabs.create {
#           url: 'http://localhost:8899/#/suggestions/' + obj.ID + '/'
#         }
#       else
#         console.log 'UNHANDLED ERROR'
#     return
#   ).bind(req)
#   details = maybeAdds[notificationId]
#   req.send JSON.stringify(URL: details.url)
#   return



# The application starts
#
chrome.notifications.getAll (notifications) ->
  for i of notifications
    if !notifications.hasOwnProperty(i)
      continue
    chrome.notifications.clear i, ->
  clientRunningCheck()
  setInterval (->
    clientRunningCheck()
    return
  ), 500



module.exports =
  appState: appState
  appSignals: appSignals
  setState: setState
