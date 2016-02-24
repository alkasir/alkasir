'use strict';

/* jshint esnext: true */
/* global require, module */

var React = require('react'),
    Router = require('react-router'),
    {Panel} = require('react-bootstrap'),
    Markdown = require('./Markdown'),
    _ = require('lodash'),
    Actions = require("../Actions"),
    Stores = require("../Stores");

var Navigation = require('react-router').Navigation;

/**
 *  Document
 */
var Document = React.createClass({

    mixins: [Router.State, Navigation],

    propTypes: {
        language: React.PropTypes.string,
        name: React.PropTypes.string,
        inline: React.PropTypes.bool,
    },


    getDefaultProps: function() {
        return {
            inline: false,
            language: "__",
        };
    },

    _storeUpdate: function(docs) {
        var docname = this.props.language + "/" + this.props.name;
        if (docs.hasOwnProperty(docname) && this.state.document !== docs[docname]) {
            this.setState({
                document: docs[docname],
            });
        }
    },

    componentDidMount: function() {
        Stores.documents.update.add(this._storeUpdate);
        Actions.getDocument(this.props.language + "/" + this.props.name);
    },

    componentWillUnmount: function() {
        Stores.documents.update.remove(this._storeUpdate);
    },

    getInitialState: function() {
        return {
            document: ""
        };
    },

    onClickHandler: function(e) {
        if (e.target.nodeName === "A") {
            if (_.startsWith(e.target.attributes.href.textContent, "http")) {
                return;
            } else if (_.startsWith(e.target.attributes.href.textContent, "mailto:")) {
                return;
            }
            e.preventDefault();
            e.stopPropagation();
            var v = e.target.attributes.href.textContent.split("/");
            if (v.length === 1) {
                this.transitionTo("document", {
                    language: this.props.language,
                    name: v[v.length - 1],
                });
            } else if (v.length >= 2) {
                this.transitionTo("document", {
                    language: v[v.length - 2],
                    name: v[v.length - 1],
                });
            }
        }
    },

    render: function() {
        return (
            <div onClick={this.onClickHandler}>
              <Markdown source={this.state.document} smallHeadings={this.props.inline} />
            </div>
        );
    }
});


module.exports = Document;
