'use strict';

/* jshint esnext: true */
/* global require, module */

import React from "react";
import { HostPatternsPage, MethodDetail, MethodTest, ServiceDetail, ServicesPage, UITest } from "./dev";
import Router, { Route, DefaultRoute, NotFoundRoute } from "react-router";
import { Home, App, Settings, NotFound, SuggestionForm, DocumentPage, SetupGuide, SuggestCurrentTab } from ".";

var routes = (
  <Route handler={App} path="/">
    <DefaultRoute name="home" handler={Home} />
    <Route name="settings" path="settings/:section?" handler={Settings} />
    <Route name="document" path="docs/:language/:name" handler={DocumentPage} />
    <Route name="suggest-current-tab" path="suggestions/" handler={SuggestCurrentTab} />
    <Route name="suggestionForm" path="suggestions/:id/" handler={SuggestionForm} />
    <Route name="services" path="services/" handler={ServicesPage} />
    <Route name="service" path="services/:id/" handler={ServiceDetail} />
    <Route name="method-test" path="methods/:id/test/" handler={MethodTest} />
    <Route name="method" path="methods/:id/" handler={MethodDetail} />
    <Route name="hostpatterns" path="hostpatterns/" handler={HostPatternsPage} />
    <Route name="setup-guide" path="setup-guide/" handler={SetupGuide} />
    <Route name="uitest" path="uitest/" handler={UITest} />
    <NotFoundRoute handler={NotFound} />
  </Route>
);

var run = function(elementID) {
        Router.run(routes, function(Handler) {
        React.render(<Handler />, document.getElementById(elementID));
    });
};

module.exports = run;
