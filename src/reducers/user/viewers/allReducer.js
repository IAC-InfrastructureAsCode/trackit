import Constants from '../../../constants';

const defaultValue = {status: true, values: []};

export default (state=defaultValue, action) => {
  switch (action.type) {
    case Constants.USER_GET_VIEWERS:
      return defaultValue;
    case Constants.USER_NEW_VIEWER_SUCCESS:
      return {status: true, values: state.values.concat([action.viewer])}
    case Constants.USER_GET_VIEWERS_SUCCESS:
      return {status: true, values: action.viewers};
    case Constants.USER_GET_VIEWERS_ERROR:
      return {status: true, error: action.error};
    case Constants.USER_GET_VIEWERS_CLEAR:
      return {status: true, values: []};
    default:
      return state;
  }
};
