/* jshint esnext: true */
/* global require, module, chrome */


var React = require('react'),
    Router = require('react-router'),
    Bootstrap = require("react-bootstrap"),
    chromeutil = require("../chromeutil"),
    Actions = require('../Actions'),
    {Panel} = Bootstrap;

import SuggestionStatus from "./SuggestionStatus";

/**
 *  SuggestionForm
 */
var SuggestCurrentTab = React.createClass({

    mixins: [Router.State, Router.Navigate],

    _update: function(item) {
        this.setState({
            URL: item.URL
        });
    },

    getInitialState: function() {
        return {
            URL: "...",
            submitted: false,
            status: "working"
        };
    },

    componentDidMount: function() {
        if (!this.state.submitted) {
            this.setState({
                submitted: true,
            });

            var self = this;
            chrome.tabs.query({
                active: true,
                currentWindow: true
            }, function(tabs) {
                if (tabs === undefined) {
                    // do nothing
                    return;
                }
                if (tabs.length > 0) {
                    var promise = Actions.suggestNew(tabs[0].url);
                    promise.otherwise((x) => {
                        var newStatus = "unknown_error";
                        if (x.reason === "ServerTimeout" ) {
                            newStatus = "central_error";
                        } else if (x.reason === "NotFound") {
                            newStatus = "not_found";
                        } else if (x.reason === "Invalid") {
                            x.details.causes.forEach(function(cause) {
                                if (cause.field === "URL" && cause.reason === "FieldValueNotSupported") {
                                    newStatus = "invalid_url";
                                }
                            });
                            newStatus = "invalid_url";
                        }
                        self.setState({
                            status: newStatus,
                        });
                    });
                }
            });
        }
    },
    render: function() {
        if (!chromeutil.enabled) {
            return (
                <Panel>Adding new sites only works from within the browser extension poup....</Panel>
            );
        }
        return (<Panel><SuggestionStatus status={this.state.status}/></Panel>);
    }

});

module.exports = SuggestCurrentTab;
