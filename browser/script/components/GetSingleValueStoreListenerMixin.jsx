/* global require, module */

// Very simple
var GetSingleValueStoreListenerMixin = {
    getInitialState: function() {
        return {
            item: {}
        };
    },

    _storeUpdate: function(item) {
        this.setState({
            item: item
        });
    },

    componentDidMount: function() {
        this.props.valueStore.reset.add(this._storeUpdate);
        this.props.action();

    },

    componentWillUnmount: function() {
        this.props.valueStore.reset.remove(this._storeUpdate);
    },
};


module.exports = GetSingleValueStoreListenerMixin;
