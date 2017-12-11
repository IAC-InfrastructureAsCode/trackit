import React, { Component } from 'react';
import { connect } from 'react-redux';
import List, {
  ListItem,
  ListItemText,
} from 'material-ui/List';
import Misc from '../../../misc';
import PropTypes from 'prop-types';
import Form from './FormComponent';
import Actions from "../../../../actions";

const Dialog = Misc.Dialog;
const DeleteConfirmation = Misc.DeleteConfirmation;

export class Item extends Component {

  constructor(props) {
    super(props);
    this.editBill = this.editBill.bind(this);
    this.deleteBill = this.deleteBill.bind(this);
  }

  editBill = (body) => {
    this.props.edit(body);
  };

  deleteBill = () => {
    this.props.delete(this.props.bill.id);
  };

  render() {

    return (
      <ListItem divider>

        <ListItemText
          disableTypography
          primary={this.props.bill.bucket + this.props.bill.path}
        />

        <div>

          <div className="inline-block">
            <Form
              account={this.props.account}
              bill={this.props.bill}
              submit={this.editBill}
            />
          </div>
          &nbsp;
          <div className="inline-block">
            <DeleteConfirmation entity="account" confirm={this.deleteBill}/>
          </div>

        </div>

      </ListItem>
    );
  }

}

Item.propTypes = {
  account: PropTypes.number.isRequired,
  bill: PropTypes.shape({
    bucket: PropTypes.string.isRequired,
    path: PropTypes.string.isRequired
  }),
  edit: PropTypes.func.isRequired,
  delete: PropTypes.func.isRequired
};

// List Component for AWS Accounts
class ListComponent extends Component {

  constructor(props) {
    super(props);
    this.getBills = this.getBills.bind(this);
    this.clearBills = this.clearBills.bind(this);
  }

  getBills() {
    this.props.getBills(this.props.account);
  }

  clearBills() {
    this.props.clearBills();
  }

  render() {
    let noBills = (!this.props.bills || !this.props.bills.length ? <div className="alert alert-warning" role="alert">No bills available</div> : "");
    let bills = (this.props.bills && this.props.bills.length ? (
      this.props.bills.map((bill, index) => (
        <Item
          key={index}
          bill={bill}
          account={this.props.account}
          edit={this.props.edit}
          delete={this.props.delete}/>
      ))
    ) : null);
    console.log(this.props.bills);
    return (
      <Dialog
        buttonName="Bills locations"
        title="Bills locations"
        secondActionName="Close"
        onOpen={this.getBills}
        onClose={this.clearBills}
      >

        <Form
          account={this.props.account}
          submit={this.props.new}
        />

        <List>
          {noBills}
          {bills}
        </List>

      </Dialog>
    );
  }

}

ListComponent.propTypes = {
  account: PropTypes.number.isRequired,
  bills: PropTypes.arrayOf(
    PropTypes.shape({
      bucket: PropTypes.string.isRequired,
      path: PropTypes.string.isRequired
    })
  ),
  getBills: PropTypes.func.isRequired,
  newBill: PropTypes.func.isRequired,
  editBill: PropTypes.func.isRequired,
  deleteBill: PropTypes.func.isRequired,
  clearBills: PropTypes.func.isRequired
};

/* istanbul ignore next */
const mapStateToProps = (state) => ({
  bills: state.aws.accounts.bills
});

/* istanbul ignore next */
const mapDispatchToProps = (dispatch) => ({
  getBills: (accountID) => {
    dispatch(Actions.AWS.Accounts.getAccountBills(accountID));
  },
  newBill: (accountID, bill) => {
    dispatch(Actions.AWS.Accounts.newAccountBill(accountID, bill))
  },
  editBill: (accountID, bill) => {
    dispatch(Actions.AWS.Accounts.editAccountBill(accountID, bill))
  },
  deleteBill: (accountID, bill) => {
    dispatch(Actions.AWS.Accounts.deleteAccountBill(accountID, bill));
  },
  clearBills: () => {
    dispatch(Actions.AWS.Accounts.clearAccountBills());
  }
});

export default connect(mapStateToProps, mapDispatchToProps)(ListComponent);
