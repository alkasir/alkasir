/* global location, localStorage */
'use strict';

import React from 'react';
import { T } from '../i18n';
import { Input, Panel } from 'react-bootstrap';
import Stores from "../Stores";
import { Document } from ".";


var handleBrowserCode = function(str) {
    var arr = str.split(":");
    if (arr.length !== 3) {
        return false;
    }
    var key = arr[0];
    var host = arr[1];
    var port = arr[2];
    if (host === "") {
        host = "localhost"
    }

    localStorage.authKey = key;
    localStorage.clientURL = "http://" + host + ":" + port;
    localStorage.browserCode = str;

    return true;

};

class BrowserCodeInput extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            value: "",
            ok: null,
        };
    }
    render() {
        var code = "";
        if (localStorage.browserCode){
            code = localStorage.browserCode;
        }
        return (
            <Input
                help={T( "ext_browser_code_field_help")}
                label={T( "browser_code")}
                onChange={this.handleChange}
                placeholder={T("ext_browser_code_field_placeholder")}
                type='text'
                defaultValue={code}
            />
        );
    }

    validationState() {
        let length = this.state.value.length;
        if (length > 10) {
            return 'success';
        } else if (length > 5) {
            return 'warning';
        } else if (length > 0) {
            return 'error';
        }
        return null;
    }

    handleChange(event) {
        var val = event.currentTarget.value;
        var ok = handleBrowserCode(val);
        this.setState({
            ok: ok,
            value: val,
        });

    }
}


export default class ChromeOptions extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            language: "",
            transListener: "",
        };
    }

    shouldComponentUpdate(nextProps, nextState) {
        if (this.state.language !== nextState.language) {
            return true;
        }
        return false;
    }

    componentDidMount() {
        var updateT = (data) => {
            this.setState({
                language: data.language,
            });
        };

        Stores.translation.reset.add(updateT);
        this._transListener = updateT;
    }

    componentWillUnmount() {
        Stores.translation.reset.remove(this._transListener);

    }

    render() {
        return (
            <Panel>
              <h1>{T("extension_options")}</h1>
              <Document name="browser_code" />
              <p>
                <BrowserCodeInput /> </p>
                <h2>{T("status")}</h2>
                <p>(NOTE only one satatus will be displayed at the same time)</p>
                <p> {T("status_ok_message")}</p>
                <p>{T("ext_browser_code_msg_client_conn_error")}</p>
                <p>{T("ext_browser_code_msg_client_newer_error")}</p>
                <p>{T("ext_browser_code_msg_client_older_error")}</p>
                <p>{T("ext_browser_code_msg_invalid_code")}</p>
            </Panel>
        );
    }
}
