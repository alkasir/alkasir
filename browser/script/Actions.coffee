W = require 'when'
request = require 'superagent'
Stores = require './Stores'
_ = require 'lodash'
{reqPromise, reqPromiseRaw} = require './net.coffee'
messagesReq = require.context "../../res/messages/", true, /^.*\.json$/
docsReq = require.context "../../res/documents/", true, /^.*\.md$/
messagesReq = require.context "../../res/messages/", true, /^.*\.json$/
settings = require "./settings"
chromeutil = require "./chromeutil"


# Send log to client
log = (level, context, msg) ->
  req = request.post("#{settings.baseURL}/api/log/browser/")
  .send({
    level: level
    context: context
    message: msg
  })
  promise = reqPromise req
  true


# returns promise
listServices = () ->
  req = request.get("#{settings.baseURL}/api/services/")
  promise = reqPromise req
  promise.then (v) ->
    Stores.service.reset.dispatch(v)
  true



# returns promise
listMethods = () ->
  req = request.get("#{settings.baseURL}/api/methods/")
  promise = reqPromise req
  promise.then (v) ->
    Stores.method.reset.dispatch(v)
  true



# test test
testMethod = (id) ->
  req = request.post("#{settings.baseURL}/api/methods/#{id}/test/")
  promise = reqPromise req
  promise.then (v) ->
    console.log v
    Stores.method.testResultAdded.dispatch(v)
  true



listHostPatterns = () ->
  req = request.get("#{settings.baseURL}/api/hostpatterns/")
  promise = reqPromise req
  promise.then (v) ->
    Stores.hostPattern.reset.dispatch(v)
  true


getStatusSummary = () ->
  req = request.get("#{settings.baseURL}/api/status/summary/")
  promise = reqPromise req
  promise.then (v) ->
    Stores.statusSummary.reset.dispatch(v)
  true


getVersion = () ->
  req = request.get("#{settings.baseURL}/api/version/")
  promise = reqPromise req
  promise.then (v) ->
    v.ok = true
    Stores.version.reset.dispatch(v)
  promise.otherwise (v) ->
    Stores.version.reset.dispatch({ok: false})
  true


getSettings = () ->
  req = request.get("#{settings.baseURL}/api/settings/")
  promise = reqPromise req
  promise.then (v) ->
    Stores.settings.reset.dispatch(v)
  promise


saveSettings = (appsettings) ->
  req = request.post("#{settings.baseURL}/api/settings/")
  .send(appsettings)
  promise = reqPromise req
  promise.then (v) ->
    # Be a bit stupid and just reload everything from the api
    getSettings()
  true



suggestNew = (url) ->
  req = request.post("#{settings.baseURL}/api/suggestions/")
  .send {
    URL: url
  }
  promise = reqPromise req
  promise



suggestSubmit = (id) ->
  req = request.post("#{settings.baseURL}/api/suggestions/#{id}/submit/")
  .send(true)
  promise = reqPromise req
  promise.then (v) ->
    console.log "successfully reported to central"
  promise


getSuggestion = (id) ->
  req = request.get("#{settings.baseURL}/api/suggestions/#{id}/")
  promise = reqPromise req
  promise.then (v) ->
    Stores.suggestions.reset.dispatch(v)
  promise


getAllSuggestions = () ->
  req = request.get("#{settings.baseURL}/api/suggestions/")
  promise = reqPromise req
  promise.then (v) ->
    Stores.suggestions.reset.dispatch(v)
  true


# validLanguages = ["en", "ar", "zh", "sv", "fa"]
validLanguages = ["en", "ar", "fa"]

getTranslation = (language) ->
  if language not in validLanguages
    language = validLanguages[0]
  lang = messagesReq "./#{language}/messages.json"
  Stores.translation.reset.dispatch({
    language: language
    messages: lang
  })
  true


getDocument = (name, fname, lng) ->
  if Stores.documents.has name
    Stores.documents.dispatch()
  fname = name
  lng = Stores.settings.item?.language
  if lng? and lng in validLanguages
    fname = name.replace "__", lng
  else
    fname = name.replace "__", "en"
  docpath =  "./" + fname + ".md"
  try
    doc = docsReq docpath
  catch err
    doc = """# document not found
    Cannot find #{docpath}
    """
  Stores.documents.set(name, doc)
  true



listNotifications = () ->
  req = request.get("#{settings.baseURL}/api/notifications/")
  promise = reqPromise req
  promise.then (v) ->
    Stores.notifications.reset.dispatch(v)
  promise.otherwise (v) -> # take no action
  true


validateConnectionString = (connectionString) ->
  req = request.post("#{settings.baseURL}/api/connections/validate/")
  .send({
    connectionString: connectionString
  })
  promise = reqPromise req
  promise


getConnections = () ->
  req = request.get("#{settings.baseURL}/api/connections/")
  promise = reqPromise req
  promise.then (v) ->
    Stores.connections.reset.dispatch(v)
  true


saveConnection = (connection) ->
  req = request.post("#{settings.baseURL}/api/connections/")
  .send(connection)
  promise = reqPromise req
  promise.then (v) ->
    # Be a bit stupid and just reload everything from the api
    getConnections()
  promise


toggleConnection = (connection) ->
  req = request.post(
    "#{settings.baseURL}/api/connections/#{connection.id}/toggle/")
  .send(connection)
  promise = reqPromise req
  promise.then (v) ->
    # Be a bit stupid and just reload everything from the api
    getConnections()
  promise


deleteConnection = (connection) ->
  req = request.del("#{settings.baseURL}/api/connections/#{connection.id}/")
  .send(connection)
  promise = reqPromise req
  promise.then (v) ->
    # Be a bit stupid and just reload everything from the api
    getConnections()
  promise


copyBrowsercode = () ->
  req = request.post("#{settings.baseURL}/api/browsercode/toclipboard/")
  .send(true)
  promise = reqPromise req
  promise.then (v) ->
    console.log("bc copied")
  promise


exportChromeExt = () ->
  req = request.post("#{settings.baseURL}/api/export/chrome-extension/")
  .send(true)
  promise = reqPromise req
  promise.then (v) ->
    console.log("chrome ext exported")
  promise


exportDebug = () ->
  url = "#{settings.baseURL}/api/export/debug/"
  if localStorage.authKey
    url =  url + "?authKey=" + localStorage.authKey
  if chromeutil.enabled
    chromeutil.gotoPage url
  else
    window.location = url


module.exports =
  listNotifications: listNotifications
  listServices: listServices
  listMethods: listMethods
  listHostPatterns: listHostPatterns
  testMethod: testMethod
  getStatusSummary: getStatusSummary
  getSettings: getSettings
  saveSettings: saveSettings
  suggestSubmit: suggestSubmit
  # listSuggestions: listSuggestions
  suggestNew: suggestNew
  getSuggestion: getSuggestion
  getAllSuggestions: getAllSuggestions
  getTranslation: getTranslation
  getDocument: getDocument
  log: log
  getConnections: getConnections
  saveConnection: saveConnection
  toggleConnection: toggleConnection
  deleteConnection: deleteConnection
  validateConnectionString: validateConnectionString
  getVersion: getVersion
  copyBrowsercode: copyBrowsercode
  exportDebug: exportDebug
  exportChromeExt: exportChromeExt
