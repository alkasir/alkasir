/* global location */

import i18n from "../i18n";
import _ from "lodash";

export var maybeChangeDirection = function(prevLang, nextLang) {
    var prevLTR = i18n.isRTL(prevLang);
    var nextLTR = i18n.isRTL(nextLang);
    if (prevLTR === nextLTR) {
        return false;
    }
    if (!prevLTR) {
        _.delay(() => {
            require("bootstrap-rtl/dist/css/bootstrap-rtl.css");
        }, 500);
        return "rtl";
    }
    location.reload();
    return "ltr"; // this never happens because reload above
};
