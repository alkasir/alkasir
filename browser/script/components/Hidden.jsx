"use strict";

import React from 'react';
import { T } from '../i18n';


class Hidden extends React.Component {
    constructor(props, state){
        super(props, state);
    }

    onClickHandler(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    render() {
        return (
            <div onClick={this.onClickHandler} style={{display: "none"}}>
                {this.props.children}
            </div>
        );
    }
}

Hidden.propTypes = {
    children: React.PropTypes.element.isRequired,
    description: React.PropTypes.any,
};

export default Hidden;
