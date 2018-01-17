import React, {Component} from 'react';
import {connect} from 'react-redux';
import {Redirect} from "react-router-dom";
import PropTypes from 'prop-types';
import Components from '../../components';
import Actions from '../../actions/index';

const Form = Components.Auth.Form;

// LoginContainer Component
export class LoginContainer extends Component {

  render() {
    if (this.props.token)
      return (<Redirect to="/"/>);
    return (<Form submit={this.props.login} loginStatus={this.props.loginStatus}/>);
  }

}

LoginContainer.propTypes = {
  login: PropTypes.func.isRequired,
  token: PropTypes.string,
  loginStatus: PropTypes.shape({
    status: PropTypes.bool,
    error: PropTypes.string
  })
};

/* istanbul ignore next */
const mapStateToProps = (state) => ({
  token: state.auth.token,
  loginStatus: state.auth.loginStatus
});

/* istanbul ignore next */
const mapDispatchToProps = (dispatch) => ({
  login: (email, password) => {
    dispatch(Actions.Auth.login(email, password))
  }
});

export default connect(mapStateToProps, mapDispatchToProps)(LoginContainer);
