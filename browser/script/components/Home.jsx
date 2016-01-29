'use strict';

/* jshint esnext: true */
/* global require, module */

var React = require('react'),
    Router = require('react-router'),
    Bootstrap = require("react-bootstrap"),
    T = require("../i18n").T,
    Stores = require("../Stores"),
    Actions = require("../Actions"),
    Notification = require("./Notification"),
    _ = require('lodash'),
    {Input, Panel, Table, Nav, NavItem, Button, Label} = Bootstrap;



var refreshNotificationsLoop = _.once(function() {
    setInterval(() => {
        Actions.listNotifications();
    }, 3000);
});


/**
 *  HomePage
 */
var HomePage = React.createClass({
    mixins: [Router.State],

    getInitialState: function() {
        return {
            notifications: [],
        };
    },

    _notificationUpdate: function(item) {
        this.setState({
            notifications: item,
        });
    },

    componentDidMount: function() {
        Stores.notifications.reset.add(this._notificationUpdate);
        Actions.listNotifications();
        refreshNotificationsLoop();
    },

    componentWillUnmount: function() {
        Stores.notifications.reset.remove(this._notificationUpdate);
    },

    render: function() {

        var items = this.state.notifications.map(
            function(v) {
                return (<Notification key={v.ID} item={v} />);
            }
        );

        return (
            <div>
              <Panel header={T( "notifications")}>
                {items}
              </Panel>
            </div>
        );
    }
});


module.exports = HomePage;
