W = require 'when'
request = require 'superagent'
signals = require 'signals'
_ = require 'lodash'
{SingleItemStore, CollectionStore, MapStore} = require './store.coffee'



class HostPatternStore extends CollectionStore
  name: "HostPatternStore"



class ServiceStore extends CollectionStore
  name: "ServiceStore"



class MethodStore extends CollectionStore
  name: "MethodStore"

  constructor: () ->
    super()
    @testResults = []
    @testResultAdded = new signals.Signal()

    @testResultAdded.add (result) =>
      @testResults.push @result
      true



class StatusSummaryStore extends SingleItemStore
  name: "StatusSummaryStore"



class StatusSummaryStore extends SingleItemStore
  name: "StatusSummaryStore"



class VersionStore extends SingleItemStore
  name: "VersionStore"



class SettingsStore extends SingleItemStore
  name: "SettingsStore"



class SuggestionsStore extends CollectionStore
  name: "SuggestionsStore"



class NotificationsStore extends CollectionStore
  name: "NotificationsStore"



class TranslationStore extends SingleItemStore
  name: "TranslationsStore"



class DocumentsStore extends MapStore
  name: "DocumentsStore"



class ConnectionsStore extends CollectionStore
  name: "ConnectionsStore"



module.exports = {
  service: new ServiceStore()
  method: new MethodStore()
  hostPattern: new HostPatternStore()
  statusSummary: new StatusSummaryStore()
  settings: new SettingsStore()
  suggestions: new SuggestionsStore()
  translation: new TranslationStore()
  connections: new ConnectionsStore()
  documents: new DocumentsStore()
  notifications: new NotificationsStore()
  version: new VersionStore()
}
