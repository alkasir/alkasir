import React from 'react';

import chromeutil from '../chromeutil';
chromeutil.enabled = true;

import ChromeOptions from "./ChromeOptions";
import DocumentApp from "./DocumentApp";

import appEntrypoint from "./appEntrypoint";


import Actions from "../Actions";
import Router, { Route, NotFoundRoute, Redirect } from "react-router";
import { NotFound, DocumentPage } from ".";
import i18n from "../i18n";

var run = () => {
    var result = false;

    var lng = "en";
    if (i18n.getCachedLangSetting()) {
        lng = i18n.getCachedLangSetting();
    }

    if (i18n.preRTL()) {
        require("bootstrap-rtl/dist/css/bootstrap-rtl.css");
    }

    if (document.getElementById("alkasir-chrome-options")) {
        React.render(<ChromeOptions />, document.getElementById("alkasir-chrome-options"));
        result = "options";
    }

    if (document.getElementById("alkasir-chrome-popup")) {
        appEntrypoint("alkasir-chrome-popup");
        result = "popup";
    }

    if (document.getElementById("alkasir-chrome-documents")) {
        var documentsRoutes = (
            <Route handler={DocumentApp} path="/">
              <Redirect path="home"
                        to="document"
                        query={{ language: "lng", name: "index" }} />
              <Route name="document"
                     path="/:language/:name"
                     handler={DocumentPage} />
              <NotFoundRoute handler={NotFound} />
            </Route>
        );
        Router.run(documentsRoutes, function(Handler) {
            React.render(<Handler />, document.getElementById("alkasir-chrome-documents"));
        });
        result = "documents";
    }
    Actions.getTranslation(lng);
    return result;
};

export default run;
