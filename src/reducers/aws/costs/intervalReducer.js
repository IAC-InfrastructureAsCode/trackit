import Constants from '../../../constants';

const defaultValue = "day";

export default (state={}, action) => {
  let intervals = Object.assign({}, state);
  switch (action.type) {
    case Constants.AWS_INSERT_COSTS_INTERVAL:
      return action.interval;
    case Constants.AWS_ADD_CHART:
      intervals[action.id] = defaultValue;
      return intervals;
    case Constants.AWS_SET_COSTS_INTERVAL:
      intervals[action.id] = action.interval;
      return intervals;
    case Constants.AWS_RESET_COSTS_INTERVAL:
      Object.keys(intervals).forEach((key) => {
        intervals[key] = defaultValue;
      });
      return intervals;
    case Constants.AWS_REMOVE_CHART:
      if (intervals.hasOwnProperty(action.id))
        delete intervals[action.id];
      return intervals;
    case Constants.AWS_CLEAR_COSTS_INTERVAL:
      return {};
    default:
      return state;
  }
};
