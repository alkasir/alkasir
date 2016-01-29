/* jshint esnext: true */
/* global chrome */

/* This file contanins delvelopment helper user interface components. Some of these might make it into a release version
but not as is. These have been written in the way that was fastest to write with little care taken to make them easy to
read or debug. */
"use strict";

var React = require('react');
var Router = require('react-router');

var { Link } = Router
var Actions = require("../../Actions")
var Stores = require("../../Stores")
var Bootstrap = require("react-bootstrap")
var { Input, Panel, Table, Button, Label} = Bootstrap

// Helper class to create
var NameLink = React.createClass({
    render: function() {

        var id, name
        var p = this.props
        if (p.link != null) {
            if (p.link.hasOwnProperty("ID")) {
                id = p.link.ID
            }
            if (p.link.hasOwnProperty("Name")) {
                name = p.link.Name
            }
        }
        if (id == null && p.id != null ) {
            id = p.id
        }
        if (name == null && p.name != null ) {
            name = p.name
        }
        if (name == null) {
            name = id
        }
        return (
            <Link to={this.props.to} params={{id: id}}>
            {name}
            </Link>
        )
    }
})

// Basic component for table views
var TableListMixin = {
    getInitialState: function() {
        return {
            items: []
        }
    },
    _update: function(items) {
        this.setState({
            items: items
        })
    },
    componentDidMount: function() {
        this._listStore.reset.add(this._update)
        this._onMount()
    },
    componentWillUnmount: function() {
        this._listStore.reset.remove(this._update)
    },
    renderTable: function() {
        var fields = this._fields || [];
        var idField = this._idField || "ID";
        var items = this.state.items;
        var renderers = this._renderers || {};

        var renderCell = function(field, row){
            if (renderers.hasOwnProperty(field)) {
                var Renderer = renderers[field]
                return (<Renderer field={field} row={row} />)
            } else {
                return row[field]
            }
        }
        var renderRow = function(row) {
            var renderTd = function(field){
                return (<td key={field}>{renderCell(field, row)}</td>)
            }
            return (<tr key={row[idField]}>{fields.map(renderTd)}</tr>)
        }
        var renderTh = function(v) {
            return (
                <th key={v}>
                    {v}
                </th>
            )
        }
        return (
            <Table>
                <thead><tr>{fields.map(renderTh)}</tr></thead>
                <tbody>{items.map(renderRow)}</tbody>
            </Table>
        )
    }
}


//
var BoolColFmt = React.createClass({
    getDefaultProps: function() {
        return {
            field: "",
            row: []
        }
    },
    render: function() {
        if (this.props.row[this.props.field]) {
            return (<Label bsStyle="success">RUNNING</Label>)
        } else {
            return (<Label bsStyle="warning">STOPPED</Label>)
        }
    }
});

var ServiceDetail = React.createClass({
    mixins: [Router.State, Router.Navigate],
    render: function() {
        var id = this.getParams().id
        return (<div> detail page service {id}</div>)
    }
});

var MethodDetail = React.createClass({
    mixins: [Router.State, Router.Navigate],
    render: function() {
        var id = this.getParams().id
        return (<div>detailed page method {id}</div>)
    }
});


var ServiceIDFieldLinkColFmt = React.createClass({
    render: function() {
        var row = this.props.row
        var field = this.props.field
        return (
            <NameLink to="service" id={row[field]} />
        )
    },
});

var MethodIDFieldLinkColFmt = React.createClass({
    render: function() {
        return (<NameLink to="method" id={this.props.row[this.props.field]} />)
    }
});

var TestTransportColFmt = React.createClass({
    getDefaultProps: function() {
        return {
            field: "",
            row: []
        }
    },
    render: function() {
        var row = this.props.row
        var field = this.props.field
        if (row['Service'].Name == "transport") {
            return (
                <Link to="method-test" params={{id:row['ID']}} >
                    <Button bsSize="xsmall">TEST</Button>
                </Link>
            )

        } else {
            return (<p/>)
        }
    }
});


var MethodsLinksListColFmt= React.createClass({
    getDefaultProps: function() {
        return {
            row: [],
            field: "",
        }
    },
    render: function() {
        var cell = this.props.row[this.props.field]
        var items
        if (cell != null) {
            items = cell.map(function(x) {
                return (<li><NameLink to="method" link={x} /></li>)
            });
        }
        return (<ul> {items} </ul>)
    },
})

var ServiceLinkstColFmt= React.createClass({
    getDefaultProps: function() {
        return {
            row: [],
            field: "",
        }
    },
    render: function() {
        var cell = this.props.row[this.props.field]
        var items
        if (cell != null) {
            items = cell.map(function(x) {
                return (<li><NameLink to="method" link={x} /></li>)
            });
        }
        return (<ul> {items} </ul>)
    },
})


var ServiceLinkColFmt= React.createClass({
    getDefaultProps: function() {
        return {
            row: [],
            field: "",
        }
    },
    render: function() {
        var cell = this.props.row[this.props.field]
        var items
        if (cell != null) {
            return (<NameLink to="service" link={cell} />)
        }
    },
})



var ServiceList = React.createClass({
    mixins: [TableListMixin],

    _fields: ["ID", "Running", "Name", "Methods" ],

    _listStore: Stores.service,

    _onMount: Actions.listServices,

    // TODO do not make component layouts flow like this. Use child components
    // instead!
    _renderers: {
        "Running": BoolColFmt,
        "ID": ServiceIDFieldLinkColFmt,
        "Methods": MethodsLinksListColFmt,
    },
    render: function() {
        return this.renderTable()
    }
})


var MethodList = React.createClass({

    mixins: [TableListMixin],

    _fields: ["ID", "Running", "Service", "Name",
              "Protocol", "BindAddr", "Test"],

    _renderers: {
        "ID": MethodIDFieldLinkColFmt,
        "Running": BoolColFmt,
        "Test": TestTransportColFmt,
        "Service": ServiceLinkColFmt,
    },

    _listStore: Stores.method,

    _onMount: Actions.listMethods,

    render: function() {
        return this.renderTable()
    }
})



var DeleteHostPatternColFmt = React.createClass({
    getDefaultProps: function() {
        return {
            field: "",
            row: []
        }
    },
    handleDelete: function(){
        /*         Actions.deleteHostPattern(this.props.row[this.props.field]) */
    },
    render: function() {
        var row = this.props.row
        var field = this.props.field
        return (<span>
                <Button bsSize="xsmall" onClick={this.handleDelete}>x</Button>
                {row[field]}
                </span>
        )
    }
});


var AddHostPatternView = React.createClass({
    getInitialState: function() {
        return {
            value: ""
        }
    },
    validate: function(){
        var len = this.state.value.length
        if (len > 1) {
            return 'success'
        } else {
            return 'error'
        }

    },
    handleChange: function(){
        this.setState({
            value: this.refs.input.getValue()
        })

    },
    onSubmit: function(e){
        e.preventDefault()
            /*         Actions.addHostPattern(this.state.value) */
    },
    render: function() {
        return (
            <form onSubmit={this.onSubmit}>
                <Input
            type="text"
            value={this.state.value}
            placeholder="..."
            label="Add item"
            bsStyle={this.validate()}
            groupClassName="group-class"
            ref="input"
            onChange={this.handleChange}
                />
                <Input type="submit"  value="Add"/>
            </form>
        )
    },
})



var HostPatternsPage = React.createClass({
    mixins: [Router.State, TableListMixin],

    _fields: ["Pattern", "Categories"],
    _listStore: Stores.hostPattern,
    _onMount: Actions.listHostPatterns,
    _idField: "Pattern",
    _renderers: {
        "Pattern": DeleteHostPatternColFmt,
    },
    render: function() {
        return (
                <div>
                <AddHostPatternView />
                {this.renderTable()}
            </div>

        )
    }
})



var MethodTestResult = React.createClass({
    getDefaultProps: function() {
        return {
            result: null
        }
    },
    render: function() {
        var report
        if (this.props.result == null) {
            report = (<div>...</div>)
        } else {
            report = (<div>{this.props.result.Message}</div>)
        }
        return (
            <Panel bsStyle="primary" header="Test results">
                {report}
            </Panel>
        )
    }
});

    
var MethodTest = React.createClass({
    mixins: [ Router.State ],
    getInitialState: function() {
        return {
            result: null
        }
    },

    _update: function(item) {
        if (item.MethodID == this.getParams().id) {
            this.setState({
                result: item
            })
        }
    },

    componentDidMount: function(rootNode) {
        Stores.method.testResultAdded.add(this._update)
        Actions.testMethod(this.getParams().id)
    },

    componentWillUnmount: function() {
        Stores.method.testResultAdded.remove(this._update)

    },
    render: function() {
        var id = this.getParams().id
        return (<div><MethodTestResult result={this.state.result} /></div>)
    }
});



var ServicesPage = React.createClass({

    mixins: [ Router.State ],

    render: function() {
        return (
            <Panel bsStyle="primary" header="Services">
                <Panel header="Services">
                    <ServiceList />
                </Panel>
                <Panel header="Methods">
                    <MethodList />
                </Panel>
            </Panel>
        )
    }
});


function generateRandAlphaNumStr(len) {

    var rdmString = "";
    for (; rdmString.length < len; rdmString += Math.random().toString(36).substr(2)) {
        return rdmString.substr(0, len);
    }

}



var bg = function(){
        return chrome.extension.getBackgroundPage().bg;
}

class UITest extends React.Component {
    maybeBlocked() {
        bg().appSignals.maybeBlocked.dispatch({
            url: generateRandAlphaNumStr(12) + ".com"
        })
    }
    transportUsed() {
        bg().appSignals.transportUsed.dispatch([generateRandAlphaNumStr(12) + ".com"])
    }

    updownClient() {
        bg().setState({
            clientConnectionOk: false
        });
        bg().setState({
            clientConnectionOk: true
        });
    }
    updownTransport() {
        bg().setState({
            transportOk: false
        });
        bg().setState({
            transportOk: true
        })
    }
    render() {
        return (
            <Panel header="test notifications">
                <p>These only works from inside the chrome extension popup</p>
                <p><Button onClick={this.transportUsed}>Alkasir being used...</Button></p>
                <p><Button onClick={this.maybeBlocked}>Detected block...</Button></p>
                <p><Button onClick={this.updownClient}>simulate client on/off</Button></p>
                <p><Button onClick={this.updownTransport}>simulate transport on/off</Button></p>
            </Panel>
        )
    }
}

module.exports = {
    HostPatternsPage: HostPatternsPage,
    MethodDetail: MethodDetail,
    MethodTest: MethodTest,
    ServiceDetail: ServiceDetail,
    ServicesPage: ServicesPage,
    UITest: UITest,
}
