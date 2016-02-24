/* global location, localStorage */

import React from 'react';
import Router, { RouteHandler } from 'react-router';
import Stores from "../Stores";
import i18n from "../i18n";

/**
 *  DocumentApp
 */
var DocumentApp = React.createClass({
    mixins: [Router.State],

    propTypes: {
        Language: React.PropTypes.string,
    },

    getInitialState: function() {
        var lng = "en";
        if (i18n.getCachedLangSetting()) {
            lng = i18n.getCachedLangSetting();
        }
        return {
            language: lng,
        };
    },

    _updateLanguage: function(settings) {
        if (settings.language !== this.state.language) {
            /* var dir = maybeChangeDirection(this.state.language, settings.language); */
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

    },

    componentWillUnmount: function() {
        Stores.settings.reset.remove(this._updateLanguage);
        Stores.translation.reset.remove(this._updateTranslation);
    },

    render: function() {
        return (
            <div>
              <RouteHandler />
            </div>
        );
    }
});


export default DocumentApp;
