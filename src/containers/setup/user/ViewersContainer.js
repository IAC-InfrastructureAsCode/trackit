import React, { Component } from 'react';
import { connect } from 'react-redux';

import Components from '../../../components';
import Actions from "../../../actions";
import PropTypes from "prop-types";
import s3square from '../../../assets/s3-square.png';

class ViewersContainer extends Component {
  componentWillMount() {
    console.log('viewers container props', this.props);
    this.props.getViewers();
  }

  render() {
    return <pre>{ JSON.stringify(this.props.viewers, null, '  ') }</pre>
  }
}

const mapStateToProps = state => {console.log(state); return {
  viewers: state.user.viewers.all,
  lastViewerCreated: state.user.viewers.lastCreated,
}};

const mapDispatchToProps = dispatch => ({
  getViewers: () => dispatch(Actions.User.getViewers()),
  createViewer: email => dispatch(Actions.User.createViewer(email)),
})

export default connect(mapStateToProps, mapDispatchToProps)(ViewersContainer);
