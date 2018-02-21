import Constants from '../../../constants';

export default (state={}, action) => {
  let costs = Object.assign({}, state);
  switch (action.type) {
    case Constants.AWS_GET_COSTS_ERROR:
      costs[action.id] = { status: true, error: action.error };
      return costs;
    case Constants.AWS_GET_COSTS:
      costs[action.id] = { status: false };
      return costs;
    case Constants.AWS_GET_COSTS_SUCCESS:
      costs[action.id] = { status: true, values: action.costs };
      return costs;
    case Constants.AWS_REMOVE_CHART:
      if (costs.hasOwnProperty(action.id))
        delete costs[action.id];
      return costs;
    default:
      return state;
  }
};
