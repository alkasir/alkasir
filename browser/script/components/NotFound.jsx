'use strict';

/* jshint esnext: true */
/* global require, module */




var React = require("react"),
    T = require("../i18n").T,
    {Panel} = require("react-bootstrap");

/**
 *  NotFound
 */
var NotFound = React.createClass({
    render: function() {
        return (
            <Panel header={T( "page_not_found")}>
              <h3>{T("page_not_found")}</h3>
            </Panel>
        );
    }
});


module.exports = NotFound;
