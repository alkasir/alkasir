'use strict';

/* jshint esnext: true */
/* global require, module */


var React = require('react'),
    Router = require('react-router'),
    Bootstrap = require("react-bootstrap"),

    Actions = require('../Actions'),
    Stores = require('../Stores'),
    T = require("../i18n").T,
    {Input, Panel, Table, Nav, NavItem, Button, Label, Well, Alert} = Bootstrap;

import SuggestionStatus from "./SuggestionStatus";
import Hidden from "./Hidden";

/**
 *  SuggestionForm
 */
var SuggestionForm = React.createClass({

    mixins: [Router.State, Router.Navigate],

    getInitialState: function() {
        return {
            URL: "...",
            submitLabel: "action_send",
            status: "initial",
            response: {},
        };
    },

    _update: function(item) {
        this.setState({
            URL: item.URL,
        });
    },

    componentDidMount: function() {
        Stores.suggestions.reset.add(this._update);
        var promise = Actions.getSuggestion(this.getParams().id);
        promise.otherwise((x) => {
            if (x.reason === "NotFound") {
                this.setState({
                    status: "not_found_client",
                });
            }
        });
    },

    componentWillUnmount: function() {
        Stores.suggestions.reset.remove(this._update);
    },


    handleSubmit: function(e) {
        e.preventDefault();
        this.setState({
            status: "working",
            submitLabel: "working",
        });

        var promise = Actions.suggestSubmit(this.getParams().id);
        promise.then( (x) => {
            this.setState({
                status: "done",
                submitLabel: "done",
            });
        });

        promise.otherwise((x) => {
            // todo: this is not the place for a translation helper
            var newStatus = "unknown_error";
            var submitLabel = "";
            if (x.reason === "ServerTimeout" ) {
                submitLabel = "action_send";
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

            var newState = {
                status: newStatus,
            }
            if (submitLabel !== "") {
                newState.submitLabel = submitLabel;
            }
            this.setState(newState);
        });
    },

    render: function() {
        var isDisabled = true;
        if (this.state.status === "initial" || this.state.status === "central_error" || this.state.status === "unknown_error") {
            isDisabled = false;
        }
        return (
            <div>
              <Panel header={T( "app_suggest_name")}>
                <form onSubmit={this.handleSubmit}>
                  <Input disabled
                         label={T( "url")}
                         type="text"
                         value={this.state.URL} />
                  <Hidden>
                  <Input disabled
                         label={T( "page_title")}
                         type="text" value="Welcome to the hello Page!" />
                  </Hidden>
                  <p>{T("suggest_explain_short")}</p>
                  <Input bsStyle="info"
                         type="submit"
                         disabled={isDisabled}
                         value={T(this.state.submitLabel)} />
                  <SuggestionStatus status={this.state.status} />
                </form>
              </Panel>
              <Hidden>
              <Panel header={T( "suggest_details_title")}>
                <p>{T("suggest_details")}:</p>
                <Well>
                  <li>redirect path: http://.... to http://... to http://... </li>
                  <li>http headers: status-code: XXX, poweredby: XXX </li>
                  <li>internet ip address: XXX</li>
                  <li>Traceroute: .../...</li>
                  <li>related hosts: images.domain.com, somecdn.akamai.something...</li>
                </Well>
              </Panel>
            </Hidden>
            </div>
        );
    },
});


module.exports = SuggestionForm;
