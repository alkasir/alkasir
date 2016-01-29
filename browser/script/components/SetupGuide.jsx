'use strict';

/* jshint esnext: true */
/* global require, module */

var React = require('react'),
    Router = require('react-router'),
    Bootstrap = require("react-bootstrap"),
    Stores = require("../Stores"),
    T = require("../i18n").T,
    Settings = require("./Settings"),
    {Input, Panel, Table, Nav, NavItem, Button, Label, Well, Alert} = Bootstrap;



/*
 *  Setup1
 */
var LanguageSetup = React.createClass({

    mixins: [Router.State],

    propTypes: {
        onDone: React.PropTypes.func,
    },
    getDefaultProps: function() {
        return {
            onDone: function() {},
        };
    },

    getInitialState: function() {
        return {
            valid: false,
        };
    },

    _settingsUpdate: function(settings) {
        this.setState({
            valid: settings.countryCode !== "__"
        });
    },

    componentDidMount: function() {
        Stores.settings.reset.add(this._settingsUpdate);

    },

    componentWillUnmount: function() {
        Stores.settings.reset.remove(this._settingsUpdate);
    },

    handleContinue: function() {
        if (this.state.valid) {
            this.props.onDone("browser");
        }
    },

    render: function() {
        return (
            <Panel header={T( "setup")}>
              <Settings section="language"
                        panels={false} />
              <Button disabled={!this.state.valid}
                      bsStyle="info"
                      onClick={this.handleContinue}>
                {T("action_continue")}
              </Button>
            </Panel>
        );
    }
});


/*
 *  Setup2
 */
var BrowserSetup = React.createClass({
    mixins: [Router.State],

    getDefaultProps: function() {
        return {
            onDone: function() {},
        };
    },

    getInitialState: function() {
        return {
            valid: true,
        };
    },

    _handleBrowserCheck: function(settings) {
        this.setState({
            valid: false, // TODO implement in extension
        });
    },

    render: function() {
        return (
            <Panel header="Setup">
              <Settings section="browser-integration" panels={false} />
            </Panel>
        );
    }
});


/**
 *  SetupGuide
 */
var SetupGuide = React.createClass({

    mixins: [Router.State],

    getInitialState: function() {
        return {
            currentStep: "lang",
        };
    },

    handleContinue: function(next) {
        this.setState({
            currentStep: next,
        });
    },

    render: function() {
        switch (this.state.currentStep) {
            case "lang":
                return (<LanguageSetup key="lang"
                                       onDone={this.handleContinue}/>);
            case "browser":
                return (<BrowserSetup key="browser"
                                      onDone={this.handleContinue}/>);
        }
        return (<div>ERROR: UNKNOWN STEP</div>);
    }
});



module.exports = SetupGuide;
