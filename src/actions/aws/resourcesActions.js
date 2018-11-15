import Constants from '../../constants';

export default {
  get: {
    EC2: (date) => ({type: Constants.AWS_RESOURCES_GET_EC2, date}),
    RDS: (date) => ({type: Constants.AWS_RESOURCES_GET_RDS, date}),
    ES: (date) => ({type: Constants.AWS_RESOURCES_GET_ES, date}),
  },
  clear: {
    EC2: () => ({type: Constants.AWS_RESOURCES_GET_EC2_CLEAR}),
    RDS: () => ({type: Constants.AWS_RESOURCES_GET_RDS_CLEAR}),
    ES: () => ({type: Constants.AWS_RESOURCES_GET_ES_CLEAR}),
  },
  setDates: (startDate, endDate) => ({
    type: Constants.AWS_RESOURCES_SET_DATES,
    dates: {startDate, endDate}
  }),
  clearDates: () => ({type: Constants.AWS_RESOURCES_RESET_DATES}),
  resetDates: () => ({type: Constants.AWS_RESOURCES_CLEAR_DATES}),
};
