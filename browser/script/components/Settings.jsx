'use strict';

/* jshint esnext: true */
/* global require, module */

var React = require('react'),
    Router = require('react-router'),
    Bootstrap = require("react-bootstrap"),
    SelectCountry = require("./SelectCountry"),
    Document = require("./Document"),

    chromeutil = require('../chromeutil'),
    Actions = require('../Actions'),
    Stores = require('../Stores'),
    T = require('../i18n').T,
    _ = require("lodash"),
    {Input, Panel, Table, Nav, NavItem, Button, Label, Modal, ProgressBar, ModalTrigger, Glyphicon, Row, Col, Alert, Well} = Bootstrap;

import Disabled from "./Disabled";
import Hidden from "./Hidden";



var ConnectionStringSetting = React.createClass({

    propTypes: {
        connection: React.PropTypes.object,
        onRequestHide: React.PropTypes.func,
    },
    getDefaultProps: function() {
        return {
            connection: {
                id: "",
                name: "",
                encoded: "",
            },
        };
    },

    getInitialState: function() {
        return {
            value: this.props.connection.encoded,
            name: "",
            valid: "initial",
            encoded: this.props.connection.encoded,
        };
    },

    handleSave: function() {

        var promise = Actions.saveConnection({
            id: this.props.connection.id,
            encoded: this.state.encoded,
        });
        promise.then((res)=> {
            this.props.onRequestHide();
        });
        promise.otherwise((res)=> {
            console.log("SAVING CONNECTION FAILED");
        });
    },

    handleChange: function(e) {
        this.setState({
            value: e.target.value,
        });
        var promise = Actions.validateConnectionString(e.target.value);
        var encoded = e.target.value;
        var c = this;
        promise.then(function(response) {
            console.log(response);
            if (response.ok) {
                c.setState({
                    encoded: encoded,
                    valid: "",
                    name: response.name,
                });
            } else {
                c.setState({
                    encoded: "",
                    valid: "warning",
                    name: ""
                });
            }

        });
    },

    render: function() {
        return (
            <Modal {...this.props}
                   bsStyle="primary"
                   title={T( "update_connection_string")}
                   animation>
              <div className="modal-body">
                <form onSubmit={this.handleSubmit}>
                  <p>{T("connection_string_help")}</p>
                  <Input type="textarea"
                         label={T( "connection_string")}
                         value={this.state.value}
                         onChange={this.handleChange}
                         bsStyle={this.state.valid} />
                  <p>{T("name")}: {this.state.name} </p>
                </form>
              </div>
              <div className="modal-footer">
                <Button onClick={this.props.onRequestHide}>{T("action_cancel")}</Button>
                <Button
                    disabled={this.state.valid !== ""}
                    onClick={this.handleSave}>{T("action_save")}</Button>
              </div>
            </Modal>
        );
    }
});

var ConnectionItemEditor = React.createClass({
    propTypes: {
        value: React.PropTypes.object,
        onRequestHide: React.PropTypes.func,
    },

    handleToggle: function(x) {
        this.props.value.disabled = !this.props.value.disabled;
        Actions.toggleConnection(this.props.value);
    },
    handleDelete: function(x) {
        Actions.deleteConnection(this.props.value);
    },


    render: function() {
        const innerRadio = (
            <input type="checkbox"
                   checked={!this.props.value.disabled}
                   onChange={this.handleToggle} />);

        if (this.props.value.protected) {
            return (
                <div>
                  <Input type="text"
                         disabled={this.props.value.protected}
                         value={this.props.value.name}
                         addonBefore={innerRadio}
                         buttonAfter={<Button disabled bsStyle="warning">
                                      {T("action_delete")}
                                      </Button>} />
                </div>
            );
        } else {
            return (
                <div>
                  <ModalTrigger
                      modal={<ConnectionStringSetting connection={this.props.value} />}>
                    <Input type="text"
                           value={this.props.value.name}
                           addonBefore={innerRadio}
                           buttonAfter={<Button
                                        onClick={this.handleDelete}
                                        bsStyle="warning">
                                        {T("action_delete")}
                                        </Button>} />
                  </ModalTrigger>
                </div>
            );
        }

    }
});

var AddConnectionEditor = React.createClass({
    propTypes: {
        value: React.PropTypes.object,
        onRequestHide: React.PropTypes.func,
    },

    handleEdit: function(x) {

    },
    render: function() {
        return (
            <div>
                <ModalTrigger modal={<ConnectionStringSetting connection={{}} />}>
                    <Button bsStyle="info">Add</Button>
            </ModalTrigger>
            </div>

        )
    }
});


var ConnectionsList = React.createClass({

    getInitialState: function() {
        return {
            connections: [],
        };
    },

    _connectionsUpdate: function(item) {
        this.setState({
            connections: item
        });
    },

    componentDidMount: function() {
        Stores.connections.reset.add(this._connectionsUpdate);
        Actions.getConnections();

    },

    componentWillUnmount: function() {
        Stores.connections.reset.remove(this._connectionsUpdate);
    },

    render: function() {

        var items = this.state.connections.map((s) => {
            return (<ConnectionItemEditor value={s} />
            );
        });


        return (
            <div>
            <div>
                {items}
            </div>
            <AddConnectionEditor />
            </div>
        )
    },



});

var MaybePanel = React.createClass({
    getDefaultProps: function() {
        return {
            "panel": true,
            "header": "",
        };
    },

    render: function() {
        if (this.props.panel) {
            return (
                <Panel header={this.props.header}>
                  {this.props.children}
                </Panel>
            );
        } else {
            var header = (<span />);
            if (this.props.header !== "") {
                header = <h3>{this.props.header}</h3>;
            }

            return (
                <div>
                  {header} {this.props.children}
                </div>
            );
        }
    },
});

/**
 *  Settings
 */
var Settings = React.createClass({

    mixins: [Router.State],

    propTypes: {
        panels: React.PropTypes.bool,
        section: React.PropTypes.string,
    },

    getDefaultProps: function() {
        return {
            section: "",
            panels: true,

        };
    },

    getInitialState: function() {
        return {
            language: "",
            languageOptions: [],
            countryCode: "",
            clientAutoUpdate: false,
            blocklistAutoUpdate: false,
            statusSummary: {},
            // notificationsDisplay: "none",
            notificationsDisplay: "block",

        };
    },

    _settingsUpdate: function(item) {
        this.setState({
            language: item.language,
            languageOptions: item.languageOptions,
            clientAutoUpdate: item.clientAutoUpdate,
            blocklistAutoUpdate: item.blocklistAutoUpdate,
            countryCode: item.countryCode,
        });
    },

    _statusUpdate: function(item) {
        this.setState({
            statusSummary: item
        });
    },

    componentDidMount: function() {
        Stores.settings.reset.add(this._settingsUpdate);
        Stores.statusSummary.reset.add(this._statusUpdate);
        Actions.getSettings();
        Actions.getStatusSummary();

    },

    componentWillUnmount: function() {
        Stores.settings.reset.remove(this._settingsUpdate);
        Stores.statusSummary.reset.remove(this._statusUpdate);
    },


    handleChangeclientAutoUpdate: function(e) {
        this.setState({
            clientAutoUpdate: e.target.checked,
        });
    },

    handleChangeblocklistAutoUpdate: function(e) {
        this.setState({
            blocklistAutoUpdate: e.target.checked,
        });
    },

    handleChangeCountry: function(e) {
        this.setState({
            countryCode: e.target.value,
        });
    },

    handleChangelanguage: function(e) {
        this.setState({
            language: e.target.value,
        });
    },

    handleToggleNotificationsDisplay: function(e) {
        this.setState({
            notificationsDisplay: this.state.notificationsDisplay === 'none' ? 'block' : 'none'
        });
    },


    componentDidUpdate: function(prevProps, prevState) {
        if (!_.isEqual(this.state, prevState)) {
            Actions.saveSettings({
                countryCode: this.state.countryCode,
                language: this.state.language,
                blocklistAutoUpdate: this.state.blocklistAutoUpdate,
            });
        }
    },

    handleCopyBrowsercode: function() {
        Actions.copyBrowsercode();
    },

    handleSaveDebug: function() {
        Actions.exportDebug();
    },

    handleSaveChromeExt: function() {
        Actions.exportChromeExt();
    },

    handleSubmit: function(e) {
        e.preventDefault();
    },

    render: function() {

        var createOption = function(option) {
            return (<option key={option} value={option}>{this(option)}</option>);
        };

        var langOpt = function(v) {
            return T("language_option_" + v);
        };

        var languageOptions = this.state.languageOptions.map(createOption, langOpt);

        var section;

        if (this.props.section !== "") {
            section = this.props.section;
        } else if (this.hasOwnProperty("getParams")) {
            section = this.getParams().section;
        }
        var sections = [];

        if (section === undefined || section === "language") {
            sections.push(<MaybePanel panel={this.props.panels}
                                      key="language"
                                      header={T( "language_and_location")}>
                  <Input type="select"
                         value={this.state.language}
                         onChange={this.handleChangelanguage}
                         addonBefore={<Glyphicon glyph="pencil" />}
                         label={T("form_label_language")} >
                  { languageOptions }
                  </Input>
                  <SelectCountry value={this.state.countryCode}
                                 onChange={this.handleChangeCountry} />
            </MaybePanel>);
        }


        if (section === undefined || section === "update") {
            sections.push(
                <MaybePanel panel={this.props.panels}
                            key="update-blocklist"
                            header={T( "block_list_updates")}>
                  <Input label={T( "last_update")} wrapperClassName="wrapper">
                  <p>{this.state.statusSummary.lastBlocklistChange}</p>
                  </Input>
                  <Input label={T( "options")}
                         wrapperClassName="wrapper">
                  <Input type="checkbox"
                         checked={this.state.blocklistAutoUpdate}
                         onChange={this.handleChangeblocklistAutoUpdate}
                         label={T( "blocklist_auto_update")} />

                  </Input>
                  <Input label={T( "actions")}
                         wrapperClassName="wrapper">
                  <Row>
                    <Col xs={5}>
                    <Button bsStyle="info">{T("action_check")}</Button>
                    </Col>
                  </Row>
                  </Input>
                </MaybePanel>
            );
        }

        if (section === undefined || section === "update") {
            sections.push(
<Disabled>
                <MaybePanel panel={this.props.panels}
                            key="update-client"
                            header={T( "application_upgrades")}>
                  <Input label={T( "status")}
                         wrapperClassName="wrapper">
                  <Row>
                    <Col xs={6}>
                    <p>{T("current_version")}</p>
                    <p>{this.state.statusSummary.alkasirVersion}</p>
                    </Col>
                    <Col xs={6}>
                    <p>{T("latest_version")}</p>
                    <p>0.2.2</p>
                    </Col>
                  </Row>
                  </Input>
                  <Input label={T( "options")}
                         wrapperClassName="wrapper">
                  <Input type="checkbox"
                         checked={this.state.clientAutoUpdate}
                         onChange={this.handleChangeclientAutoUpdate}
                         label={T( "application_auto_upgrade")} />
                  </Input>
                  <Input label={T( "actions")}
                         wrapperClassName="wrapper">
                  <Row>
                    <Col xs={5}>
                    <Button bsStyle="info">{T("action_check")}</Button>
                    </Col>
                    <Col xs={5}>
                    <Button bsStyle="info" disabled>{T("action_upgrade")}</Button>
                    </Col>
                  </Row>
                  </Input>
                </MaybePanel>
</Disabled>
            );
        }


        if ((!chromeutil.enablede && section === undefined) || section === "browser-integration") {
            var msg = T("browser_integration_error");
            if (this.state.statusSummary.browserOk) {
                msg = T("browser_integration_chrome_success");
            }
            sections.push(
                <MaybePanel panel={this.props.panels} key="browser-integration" header={T( "browser_integration")}>

                  <Input label={T( "browser_code")} wrapperClassName="wrapper">
                  <p>{T("browser_code_explain")}</p>
                  <Input type="text"
                         disabled value="*hidden*"
                         buttonBefore={<Button
                                           bsStyle="info"
                                           onClick={this.handleCopyBrowsercode}
                                                    >{T("copy_to_clipboard")}</Button>} />
                  </Input>

                  <Well>
                    <Document name="browser_extension_install" inline />
                  </Well>
<Disabled>
                  <Input label={T( "status")}
                         wrapperClassName="wrapper">
                  <Alert bsStyle="success">
                      <Glyphicon style={{ verticalAlign: "middle" }}
                                 glyph="ok" /> &nbsp;{msg}
                  </Alert>
                  <Alert bsStyle="warning">
                    <h4>
                      <Glyphicon style={{verticalAlign: "middle"}}
                                 glyph="exclamation-sign" />&nbsp;{T("browser_integration_error")}
                    </h4>
                  </Alert>
                  </Input>
</Disabled>
                </MaybePanel>
            );
        }

        if (section === undefined || section === "notifications") {
            var notifyitems = [
                {
                    header: 'general',
                    items: ['client_error', 'network_error']
                },
                {
                    header: 'browser_extension',
                    items: ['connected', 'client_not_available', 'error']
                },
                {
                    header: 'blocklist_update',
                    items: ['success', 'error']
                },
                {
                    header: 'upgrade',
                    items: ['new_version_available', 'upgrading', 'error']
                },
                {
                    header: 'central',
                    items: ['error']
                },
                {
                    header: 'transport',
                    items: ['connected', 'error', "conn_opened"]
                },
                {
                    header: 'suggest',
                    items: ['detected']
                },
            ];

            var notifysections = notifyitems.map((s) => {
                return (
                    <div>
                        <h3>{T("notify_section_" + s.header)}</h3>
                        {s.items.map((i) => {
                            return (
                                <Row>
                                <Col xs={6}>{T("notify_option_" + s.header + "_" + i)}</Col>
                                <Col xs={3}><Input type='checkbox' label={T( "notify_type_desktop")} /> </Col>
                                <Col xs={3}><Input type="checkbox" label={T( "notify_type_icon")} /> </Col>
                                </Row>
                            );
                         })}
                    </div>
                );
            });

            // var msgs = {};
            // notifyitems.map( (s) => {
            //     msgs["notify_section_" + s.header] = {
            //         "message": ""
            //     };
            //     s.items.map((i) => {
            //         msgs["notify_option_" + s.header + "_" + i] = {
            //             message: ""
            //         }

            //     })
            // });
            // console.log(JSON.stringify(msgs))



            sections.push(
<Disabled>
                <MaybePanel panel={this.props.panels}
                            key="notifications"
                            header={T( "notification_settings")}>
                  <p>{T("notification_settings_explain_short")}</p>
                  <Button bsStyle="info">{T("restore_default_settings")}</Button>
                  <Button bsStyle="info"
                          onClick={this.handleToggleNotificationsDisplay}>
                      {T("toggle_detailed_settings")}
                  </Button>
                  <div style={{ "display": this.state.notificationsDisplay }}>
                    {notifysections}
                </div>
                    </MaybePanel>
</Disabled>
            );
        }

        if (section === undefined || section === "advanced") {
            sections.push(
                <MaybePanel panel={this.props.panels} key="advanced" header={T( "advanced_settings")}>

<Hidden>
                  <Input label={T( "additional_features")} wrapperClassName="wrapper">
                  <Input type="checkbox"
                         checked={this.state.blocklistAutoUpdate}
                         onChange={this.handleChangeblocklistAutoUpdate}
                         label={T( "allow_extended_analysis")} />
                  </Input>
</Hidden>

                <h3>{T("transport_connections_title")}</h3>
                <p>{T("transport_connections_message")}</p>
                <ConnectionsList />

                <h3>{T("export_browser_extension_title")}</h3>
                <p>{T('export_browser_extension_message')}</p>
                <Button onClick={this.handleSaveChromeExt} bsStyle="info">{T("action_save")}</Button>


                <h3>{T("export_debuginfo_title")}</h3>
                <p>{T('export_debuginfo_message')}</p>
                <Button onClick={this.handleSaveDebug} bsStyle="warning">{T("action_save")}</Button>


                </MaybePanel>
            );
        }
        return (
            <form onSubmit={this.handleSubmit}>
              {sections}
            </form>
        );
    },
});


module.exports = Settings;
