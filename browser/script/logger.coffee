Actions = (require "./Actions")


# # Logger
class Logger
  constructor: (context) ->
    @context = context

  log: (level, message) =>
    Actions.log level, @context, message

newLogger = (context) ->
  (new Logger(context)).log

            

module.exports = newLogger
