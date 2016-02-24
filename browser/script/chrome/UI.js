require("../../style/main.less");

import settings from "../settings";
settings.baseURL = "http://localhost:8899";

import Actions from "../Actions";
import i18n from "../i18n";
if (i18n.preRTL()) {
    require("bootstrap-rtl/dist/css/bootstrap-rtl.css");
}


Actions.getSettings();

import chromeEntrypoint from "../components/chromeEntrypoint";

chromeEntrypoint();
