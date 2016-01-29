/* globals require, module, __dirname */

var c = require("./webpack.config");

c.entry = {
        ui: "./browser/script/chrome/UI",
        background: "./browser/script/chrome/backgroundEntrypoint"
};
c.output.path = __dirname + '/browser/chrome-ext/src/javascripts/';
c.output.publicPath = '/javascripts/';

module.exports = c;
