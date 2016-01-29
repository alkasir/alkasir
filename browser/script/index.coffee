
require "../style/main.less"

i18n = require "./i18n"
if i18n.preRTL()
  require "bootstrap-rtl/dist/css/bootstrap-rtl.css"

require 'when/monitor/console'
routes = require "./components/appEntrypoint"
log = (require "./logger")("index")

Actions = require "./Actions"

Actions.getSettings()
routes("content")
