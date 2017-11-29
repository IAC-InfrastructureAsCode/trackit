import React, { Component } from 'react';
import { connect } from 'react-redux';

import Components from '../../../components';
import Actions from "../../../actions";
import PropTypes from "prop-types";
import s3square from '../../../assets/s3-square.png';

const List = Components.AWS.Accounts.List;
const Form = Components.AWS.Accounts.Form;
const Panel = Components.Misc.Panel;

// Accounts Container for AWS Accounts
export class AccountsContainer extends Component {

  componentWillMount() {
    this.props.getAccounts();
    this.props.newExternal();
  }

  render() {
    return (
      <Panel>

        <div>

          <h3 className="white-box-title no-padding inline-block">
            <img className="white-box-title-icon" src={s3square} alt="AWS square logo"/>
            AWS Accounts
          </h3>

          <div className="inline-block pull-right">
            <Form
              submit={this.props.accountActions.new}
              external={this.props.external}
            />
          </div>

        </div>

        <List
          accounts={this.props.accounts}
          accountActions={this.props.accountActions}
          billActions={this.props.billActions}
        />

      </Panel>
    );
  }

}

AccountsContainer.propTypes = {
  accounts: PropTypes.arrayOf(
    PropTypes.shape({
      id: PropTypes.number.isRequired,
      roleArn: PropTypes.string.isRequired,
      pretty: PropTypes.string,
      bills: PropTypes.arrayOf(
        PropTypes.shape({
          bucket: PropTypes.string.isRequired,
          path: PropTypes.string.isRequired
        })
      )
    })
  ),
  external: PropTypes.string,
  getAccounts: PropTypes.func.isRequired,
  accountActions: PropTypes.shape({
    new: PropTypes.func.isRequired,
    edit: PropTypes.func.isRequired,
    delete: PropTypes.func.isRequired,
  }).isRequired,
  billActions: PropTypes.shape({
    new: PropTypes.func.isRequired,
    edit: PropTypes.func.isRequired,
    delete: PropTypes.func.isRequired,
  }).isRequired,
  newExternal: PropTypes.func.isRequired
};

/* istanbul ignore next */
const mapStateToProps = (state) => ({
  accounts: state.aws.accounts.all,
  external: state.aws.accounts.external
});

/* istanbul ignore next */
const mapDispatchToProps = (dispatch) => ({
  getAccounts: () => {
    dispatch(Actions.AWS.Accounts.getAccounts())
  },
  accountActions: {
    new: (account) => {
      dispatch(Actions.AWS.Accounts.newAccount(account))
    },
    edit: (account) => {
      dispatch(Actions.AWS.Accounts.editAccount(account))
    },
    delete: (accountID) => {
      dispatch(Actions.AWS.Accounts.deleteAccount(accountID));
    },
  },
  billActions: {
    new: (accountID, bill) => {
      dispatch(Actions.AWS.Accounts.newAccountBill(accountID, bill))
    },
    edit: (accountID, bill) => {
      dispatch(Actions.AWS.Accounts.editAccountBill(accountID, bill))
    },
    delete: (accountID, bill) => {
      dispatch(Actions.AWS.Accounts.deleteAccountBill(accountID, bill));
    },
  },
  newExternal: () => {
    dispatch(Actions.AWS.Accounts.newExternal())
  },
});

export default connect(mapStateToProps, mapDispatchToProps)(AccountsContainer);
