/* jshint esnext: true */

var React = require('react'),
    Router = require('react-router'),
    Bootstrap = require("react-bootstrap"),

    Constants = require("../constants"),
    T = require("../i18n").T,
    {Input, Glyphicon} = Bootstrap;


/**
 *  countries returns list of all countries as option tags
 */
var countries = function(currentValue) {
    var countrieslist = [],
        c = Constants.countries,
        hasSelected = currentValue !== "__";

    for (var p in c) {
        if (c.hasOwnProperty(p)) {
            var v = c[p];
            if (hasSelected && p === "__") {
                continue;
            }
            countrieslist.push(<option value={p} key={p}>{v}</option>);
        }
    }
    return countrieslist;
};


/**
 *  SelectCountry
 */
var SelectCountry = React.createClass({
    mixins: [Router.State],

    propTypes: {
        value: React.PropTypes.any,
        onChange: React.PropTypes.func,
    },

    isValid: function() {
        if (this.props.value === "__") {
            return "warning";
        }
        return "";
    },

    render: function() {
        var countrylist = countries(this.props.value);
        return (
            <Input type="select"
                   label={T( "form_label_country")}
                   value={this.props.value}
                   bsStyle={this.isValid()}
                   onChange={this.props.onChange}
                   addonBefore={<Glyphicon glyph="globe" />} >
            {countrylist}
            </Input>
        );
    },
});

module.exports = SelectCountry;
