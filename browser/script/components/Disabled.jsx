"use strict";

import React from 'react';
import { T } from '../i18n';


class Disabeled extends React.Component {
    constructor(props, state){
        super(props, state);
    }

    onClickHandler(e) {
        /* e.preventDefault(); */
        /* e.stopPropagation(); */
    }

    render() {
        return (
            <div onClick={this.onClickHandler}
                 style={{position: "relative"}}>
                <div style={{ position: "absolute",
                              top: "1em",
                              left: "1em",
                              color: "white",
                              fontWeight: 900,
                              padding: "0.5em",
                              zIndex: 101,
                              backgroundColor: "red",
                              opacity: 0.8,
                            }}>
                    DISABLED
                </div>
                <div style={{ position: "absolute",
                              backgroundColor: "red",
                              zIndex: 100,
                              height: "100%",
                              width: "100%",
                              opacity: 0.5,
                            }} />
                {this.props.children}
            </div>
        );
    }
}

Disabeled.propTypes = {
    children: React.PropTypes.element.isRequired,
    description: React.PropTypes.any,
};

export default Disabeled;
