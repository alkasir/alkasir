'use strict';

/* jshint esnext: true */
/* global require, module */

var React = require('react'),
    Router = require('react-router'),
    {Panel} = require('react-bootstrap'),
    Document = require('./Document'),
    T = require("../i18n").T;

var Navigation = require('react-router').Navigation;


/**
 *  DocumentPage
 */
var DocumentPage = React.createClass({

    mixins: [Router.State, Navigation],

    getDefaultProps: function() {
        return {
            language: "__",
            name: "404",
        };
    },

    render: function() {
        var params = this.getParams();
        return (
            <Panel header={T( "app_docs_name")}>
              <Document key={params.language + "/" + params.name}
                        language={params.language}
                        name={params.name} />
            </Panel>
        );
    }
});


module.exports = DocumentPage;
