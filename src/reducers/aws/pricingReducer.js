import Constants from '../../constants';

export default (state=[], action) => {
  switch (action.type) {
    case Constants.AWS_GET_PRICING_SUCCESS:
      return action.pricing;
    default:
      return state;
  }
};
