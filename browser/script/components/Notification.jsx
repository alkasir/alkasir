/* jshint esnext: true */
/* global require, module, chrome */

var React = require('react'),
    Bootstrap = require("react-bootstrap"),

    Router = require('react-router'),
    {Link} = Router,

    {Alert} = Bootstrap,
    T = require("../i18n").T;


/**
 *  Alert
 */
var Notification = React.createClass({

    propTypes: {
        item: React.PropTypes.any,
    },

    componentWillMount: function() {

        if (this.props.item.title === "suggest_this_title" && chrome != null && chrome.tabs != null) {
            chrome.tabs.query({
                active: true,
                currentWindow: true
            }, (tabs) => {
                if (tabs === undefined) {
                    // do nothing
                    return;
                }
                if (tabs.length > 0) {
                    this.setState({
                        URL: tabs[0].url,
                    })

                }
            });
        }
    },

    getInitialState: function() {
        return {
            URL: "",

        };
    },


    getDefaultProps: function() {
        return {
            item: {
                level: "info",
                title: "[header]",
                message: "[message]",
                actions: []
            },
        };
    },

    render: function() {
        // var level = this.props.item.level;
        var hidden = false;
        var actions = [],
            count = 0;
        if (this.props.item.actions !== null) {
            actions = this.props.item.actions.map(function(a) {
                var links = [];
                if (count > 0) {
                    links.push(<span key={count}>, </span>);
                }
                links.push(<Link key={a.route + a.name} to={a.route} className="alert-link">{T(a.name)}</Link>);
                count++;
                return (<span key={a.route + a.name}>{links}</span>);
            });
        }
        if (actions.length > 0) {
            actions = (<p key="actions">{actions}.</p>);
        } else {
            actions = (<span key="actions" />);
        }

        var message = (<span key="msg" />);
        if (this.props.item.message !== "") {
            if (this.props.item.message == "suggest_this_message") {
                if (this.state.URL === "") {
                    hidden = true;
                } else {
                    message = (<p key="msg">{T(this.props.item.message, {url: this.state.URL})}</p>);
                }
            } else {
                message = (<p key="msg">{T(this.props.item.message)}</p>);
            }
        }
        if (hidden) {
            return (<div />);
        } else {
            return (
                <Alert key={this.props.item.ID} bsStyle={this.props.item.level}>
                    <h4>{T(this.props.item.title)}</h4> {message} {actions}
                </Alert>);
        }
    }
});


module.exports = Notification;
