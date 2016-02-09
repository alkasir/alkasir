W = require 'when'
request = require 'superagent'

addAuthKey = (req) ->
  if localStorage.authKey
    req.set("Authorization", "Bearer " + localStorage.authKey)
  req

reqPromise = (request) ->
  d = W.defer()
  request = addAuthKey(request)
  request.end (err, res) ->
    if err?
      if err.type is "text/plain"
        if err.response?.text?
          d.reject err.response.text
        else
          d.reject err
      else
        if err.response?.body?
          d.reject err.response.body
        else
          d.reject err
      console.log "REQUEST ERR", request.url, err
    else
      if res.type is "text/plain"
        d.resolve res.text
      else
        d.resolve res.body

  d.promise

reqPromiseRaw = (request) ->
  d = W.defer()
  request = addAuthKey(request)
  request.end (err, res) ->
    if err?
      console.log "FAILES REQUEST", err
      d.reject err
    else
      d.resolve res

  d.promise


module.exports = {
  reqPromise: reqPromise
  reqPromiseRaw: reqPromiseRaw
}
