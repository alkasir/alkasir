W = require 'when'
request = require 'superagent'
signals = require 'signals'
_ = require 'lodash'


class SingleItemStore
  constructor: () ->
    @item = {}
    @reset = new signals.Signal()
    @change = new signals.Signal()

    @reset.add (item) =>
      @item = item
      @change.dispatch(item)
      true



# List store
class CollectionStore
  constructor: () ->
    @items = []
    @itemAdded = new signals.Signal()
    @reset = new signals.Signal()
    @change = new signals.Signal()

    @reset.add (items) =>
      @items = items
      @change.dispatch(items)
      true


# Key/value store
class MapStore
  constructor: () ->
    @items = {}
    @update = new signals.Signal()

  set: (k, v) =>
    @items[k] = v
    @update.dispatch @items

  has: (k) =>
    @items.hasOwnProperty k

  dispatch: =>
    @update.dispatch(@items)



module.exports = {
  SingleItemStore: SingleItemStore
  CollectionStore: CollectionStore
  MapStore: MapStore
}
