import { combineReducers } from 'redux';
import pricing from './pricingReducer';
import accounts from './accounts';

export default combineReducers({
  pricing,
  accounts,
});
