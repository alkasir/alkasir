/* jshint esnext: true */
/* global require, module */


var React = require('react'),
    Bootstrap = require("react-bootstrap"),


    T = require("../i18n").T,
    {Alert} = Bootstrap;

// var statusToI18n = function(x) {
//     var newStatus = "unknown_error";
//     if (x.reason === "ServerTimeout" ) {
//         newStatus = "central_error";
//     } else if (x.reason === "NotFound") {
//         newStatus = "not_found";
//     } else if (x.reason === "Invalid") {
//         x.details.causes.forEach(function(cause) {
//             if (cause.field === "URL" && cause.reason === "FieldValueNotSupported") {
//                 newStatus = "invalid_url";
//             }
//         });
//         // TODO ???
//         newStatus = "invalid_url";
//     }
//     return newStatus;
// }


var SuggestionStatus = React.createClass({
    getDefaultProps: function() {
        return {
            status: "initial",
            response: {},
        };
    },



    render: function() {
        var body;
        var status = this.props.status;
        // var responseStatus={};
        // if (this.props.response !== {}) {
            // responseStatus = statusToI18n(this.props.response)
        // }
        switch (status){
            case "initial":
                body = (
                    <span />
                );
                break;

            case "working":
                body = (
                    <Alert bsStyle="info">
                      <h4>{T("working")}</h4>
                      <p>{T("working")}</p>
                    </Alert>
                );
                break;

            case "done":
                body = (
                    <Alert bsStyle="success">
                      <h4>{T("done")}</h4>
                      <p>{T("suggest_done_message")}</p>
                    </Alert>
                );
                break;

            case "central_error":
                body = (
                    <Alert bsStyle="danger">
                      <h4>{T("central_error_title")}</h4>
                      <p>{T("central_error_message")}</p>
                    </Alert>
                );
                break;

            case "not_found":
            case "not_found_client":
                body = (
                    <Alert bsStyle="danger">
                      <h4>{T("not_found")}</h4>
                      <p>{T("suggest_not_found")}</p>
                    </Alert>
                );
                break;

            case "invalid_url":
                body = (
                    <Alert bsStyle="warning">
                      <h4>{T("suggest_denied_title")}</h4>
                      <p>{T("suggest_denied_message")}</p>
                    </Alert>
                );
                break;

            default:
                body = (
                    <Alert bsStyle="danger">
                      <h4>{T("unknown_error_title")}</h4>
                      <p>{T("unknown_error_message")}</p>
                    </Alert>
                );
        }

        if (this.props.status === "laskjdalksdjald") {
            body = (
                <span>
                  <Alert bsStyle="danger">
                    <h4>{T("network_timeout_error_title")}</h4>
                    <p>{T("network_timeout_error_message")}</p>
                  </Alert>
                </span>
            );
        }
        return body;

    },
});


module.exports = SuggestionStatus;
