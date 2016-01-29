Actions = require "./Actions"
Stores = require "./Stores"
W = require 'when'
_ = require 'lodash'


# Has translation files been loaded?
loaded = W.defer()

currentLanguage = ""
currentTrans = {}

ltrLangs = {
  "ar": true,
  "fa": true,
}

validLangs = {
  "en": true,
  "sv": true,
  "zh": true,
  "ar": true,
  "fa": true,
}

isRTL = (lng) ->
  _.has  ltrLangs, lng

getCachedLangSetting = () ->
  if localStorage["langSetting"] != undefined
    lng = localStorage["langSetting"]
    if _.has validLangs, lng
      return lng
  return false


setCachedLangSetting = (lng) ->
  if _.has validLangs, lng
    localStorage["langSetting"] = lng
    return true
  false

preRTL = () ->
  isRTL(getCachedLangSetting())

languageChangeHandler = (settings) ->
  if settings.language != currentLanguage
    currentLanguage = settings.language
    setCachedLangSetting settings.language
    Actions.getTranslation(currentLanguage)

Stores.settings.reset.add languageChangeHandler

translationUpdateHandler = (translation) ->
  currentTrans = translation.messages
  loaded.resolve true

Stores.translation.reset.add translationUpdateHandler

# Translate
T = (key, ctx = {}) ->
  res = "[#{key}]"
  if currentTrans.hasOwnProperty(key)
    ct = currentTrans[key]
    if ct.message is ""
      return res
    res = ct.message
    if ct.placeholders?
      for p of ct.placeholders
        if ctx.hasOwnProperty p
          res = res.replace('$' + p.toUpperCase() + "$", ctx[p])
  return res

module.exports = {
  load: loaded.promise
  T: T
  isRTL: isRTL
  preRTL: preRTL
  getCachedLangSetting: getCachedLangSetting
}
