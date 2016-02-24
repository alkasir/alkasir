/* global location, localStorage */

import React from 'react';
import Router, { RouteHandler } from 'react-router';
// import { DropdownButton, Panel, Nav, Navbar } from "react-bootstrap";
import { Panel, Nav, Navbar } from "react-bootstrap";
import { NavItemLink } from 'react-router-bootstrap';
import Stores from "../Stores";
import Actions from "../Actions";
import _ from "lodash";
import i18n, { T } from "../i18n";
import { maybeChangeDirection } from "./utils";
import Notification from "./Notification";


var getVersionLoop = _.once(function() {
    Actions.getVersion();
    setInterval(() => {
        Actions.getVersion();
    }, 3000);
});

var CLIENT_API_VERSION = 2;

/**
 *  App
 */
var App = React.createClass({
    mixins: [Router.State],
    getInitialState: function() {
        var lng = "en";
        if (i18n.getCachedLangSetting()) {
            lng = i18n.getCachedLangSetting();
        }
        return {
            language: lng,
            dirclass: "",
            clientConnected: true,
            apiVersion: CLIENT_API_VERSION,
        };
    },

    _updateVersion: function(data) {
        if (this.state.clientConnected !== data.ok || this.state.apiVersion !== data.apiVersion) {
            this.setState({
                clientConnected: data.ok,
                apiVersion: data.apiVersion
            });
        }
    },

    _updateLanguage: function(settings) {
        if (settings.language !== this.state.language) {
            var dir = maybeChangeDirection(this.state.language, settings.language);
            if (dir) {
                this.setState({
                    dirclss: dir
                });
            }
            this.setState({
                language: settings.language
            });
            this.forceUpdate();
        }
    },

    _updateTranslation: function() {
        this.forceUpdate();
    },


    componentDidMount: function() {
        Stores.settings.reset.add(this._updateLanguage);
        Stores.translation.reset.add(this._updateTranslation);
        Stores.version.reset.add(this._updateVersion);
        getVersionLoop();
    },

    componentWillUnmount: function() {
        Stores.settings.reset.remove(this._updateLanguage);
        Stores.translation.reset.remove(this._updateTranslation);
    },

    render: function() {
        if (!this.getPath().startsWith("/docs/")){
            if (!this.state.clientConnected) {
                var item = {
                    level: "warning",
                    title: "extension_client_not_available_title",
                    message: "extension_client_not_avialable_open_browser_options",
                    actions: [{
                        name: "action_help",
                        route: "/docs/__/index"
                    }]
                };
                return (
                    <div className={this.state.dirclass}>
                      <Panel>
                        <Notification item={item} />
                      </Panel>
                    </div>
                );
            } else if (this.state.apiVersion !== CLIENT_API_VERSION) {
                var msg = '';
                if (this.state.apiVersion < CLIENT_API_VERSION) {
                    msg = "ext_browser_code_msg_client_older_error";
                } else {
                    msg = "ext_browser_code_msg_client_newer_error";
                }
                var apiItem = {
                    level: "warning",
                    title: "extension_client_not_available_title",
                    message: msg,
                    actions: [{
                        name: "action_help",
                        route: "/docs/__/index"
                    }]
                };
                return (
                    <div className={this.state.dirclass}>
                      <Panel>
                        <Notification item={apiItem} />
                      </Panel>
                    </div>
                );
            }
        }
        var dev = (<span/>);
        // dev=(
        //     <DropdownButton eventKey={3} title="Dev">
        //             <NavItemLink to="uitest">uitest</NavItemLink>
        //             <NavItemLink to="services">services</NavItemLink>
        //             <NavItemLink to="hostpatterns">hostpatterns</NavItemLink>
        //           </DropdownButton>
        // );
        return (
            <div className={this.state.dirclass}>
              <Navbar>
                <Nav>
                  <NavItemLink to="home">{T("app_home_name")}</NavItemLink>
                  <NavItemLink to="settings">{T("app_settings_name")}</NavItemLink>
                  <NavItemLink to="document" params={{ language: "__", name: "index" }}>
                    {T("app_docs_name")}
                  </NavItemLink>
                  {dev}
                </Nav>
              </Navbar>
              <RouteHandler />
            </div>
        );
    }
});


export default App;
