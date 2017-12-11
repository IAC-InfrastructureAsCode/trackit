import Constants from '../../../constants';

export default (state=[], action) => {
  switch (action.type) {
    case Constants.AWS_GET_ACCOUNTS_SUCCESS:
      return action.accounts;
    case Constants.AWS_GET_ACCOUNTS_ERROR:
    case Constants.AWS_GET_ACCOUNTS_CLEAR:
      return [];
    default:
      return state;
  }
};
