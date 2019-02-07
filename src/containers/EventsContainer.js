import React, { Component } from 'react';
import moment from 'moment';
import PropTypes from 'prop-types';
import { connect } from 'react-redux';
import Components from '../components';
import Actions from '../actions';
import Spinner from "react-spinkit";

const TimerangeSelector = Components.Misc.TimerangeSelector;
const Filters = Components.Events.Filters.List;
const Popover = Components.Misc.Popover;

// EventsContainer Component
class EventsContainer extends Component {
  constructor() {
    super();
    this.state = {
      showHidden : false,
    }
  }

  componentDidMount() {
    if (this.props.dates) {
      const dates = this.props.dates;
      this.props.getData(dates.startDate, dates.endDate);
    }
  }

  componentWillReceiveProps(nextProps) {
    if (nextProps.dates && (this.props.dates !== nextProps.dates || this.props.accounts !== nextProps.accounts || this.props.filters !== nextProps.filters))
      nextProps.getData(nextProps.dates.startDate, nextProps.dates.endDate);
  }

  toggleHidden() {
    this.setState({ showHidden : !this.state.showHidden});
  }

  formatEvents(events, snoozed) {
    const abnormalsList = [];
    let hiddenEvents = 0;

    Object.keys(events).forEach((account) => {
      Object.keys(events[account]).forEach((key) => {
        const event = events[account][key];
        const abnormals = event.filter((item) => (snoozed ? item.abnormal : item.abnormal && !item.snoozed && !item.filtered));
        hiddenEvents += event.filter((item) => (item.snoozed || item.filtered)).length;
        abnormals.forEach((element) => {
          abnormalsList.push({element, key, event});
        });
      });
    });

    abnormalsList.sort((a, b) => ((moment(a.element.date).isBefore(b.element.date)) ? 1 : -1));

    const nodes = abnormalsList.map((abnormal) => {
      const element = abnormal.element;
      const key = abnormal.key;
      const dataSet = abnormal.event;
      return (
        <div key={`${element.date}-${key}`}>
          <Components.Events.EventPanel
            dataSet={dataSet}
            abnormalElement={element}
            service={key}
            snoozeFunc={this.props.snoozeEvent}
            unsnoozeFunc={this.props.unsnoozeEvent}
          />
        </div>
      );
    });

    return {nodes, hiddenEvents};
  }

  render() {
    const loading = (!this.props.values.status ? (<Spinner className="spinner" name='circle'/>) : null);

    const error = (this.props.values.hasOwnProperty("error") && this.props.values.error ? ` (${this.props.values.error.message})` : null);
    const emptyTimerange = (this.props.values.status && this.props.values.hasOwnProperty("values") && this.props.values.values && !Object.keys(this.props.values.values).length ? ` (Timerange not processed yet)` : null);
    const noEvents = (this.props.values.status && (error || emptyTimerange) ? <div className="alert alert-warning" role="alert">No event available{error || emptyTimerange}</div> : null);

    const timerange = (this.props.dates ?  (
      <TimerangeSelector
        startDate={this.props.dates.startDate}
        endDate={this.props.dates.endDate}
        setDatesFunc={this.props.setDates}
      />
    ) : null);

    let events = [];
    let hiddenEvents = 0;
    if (this.props.values && this.props.values.status && this.props.values.values) {
      const data = this.formatEvents(this.props.values.values, this.state.showHidden);
      events = data.nodes;
      hiddenEvents = data.hiddenEvents;
    }

    const emptyEvents = (!events.length && !loading && !noEvents ? (
      <div className="alert alert-success" role="alert">No events found for this timerange</div>
    ) : null);

    const spinnerAndError = (loading || noEvents || emptyEvents ? (
      <div className="white-box">
        {loading}
        {noEvents}
        {emptyEvents}
      </div>
    ) : null);

    const toggleHiddenButton = (hiddenEvents ? (
      <div className="inline-block">
        <Popover
          icon={
            <button className={"btn btn-default inline-block " + (this.state.showHidden ? "enabled" : "")} onClick={this.toggleHidden.bind(this)}>
              <i className={"fa fa-eye" + (!this.state.showHidden ? "-slash" : "")}/>
              &nbsp;
              {hiddenEvents} hidden events
            </button>
          }
          tooltip={"Click this to " + (this.state.showHidden ? "hide" : "see") + " snoozed / filtered events"}
          placement="top"
        />
      </div>
    ) : (
      <div className="btn btn-default inline-block disabled">
        No hidden events
      </div>
    ));

    return (
      <div>
        <div className="row">
          <div className="col-md-12">
            <div className="white-box">
              <h3 className="white-box-title no-padding inline-block">
                <i className="fa fa-exclamation-triangle"/>
                &nbsp;
                Events
              </h3>
              <div className="inline-block pull-right">
                {toggleHiddenButton}
                &nbsp;
                <div className="inline-block">
                  <Filters
                    filters={this.props.filters}
                    filterEdition={this.props.setFilters}
                    actions={this.props.filtersActions}
                  />
                </div>
                &nbsp;
                {timerange}
              </div>
            </div>
          </div>
        </div>
        {spinnerAndError}
        {events}
      </div>
    );
  }

}

EventsContainer.propTypes = {
  dates: PropTypes.object.isRequired,
  values: PropTypes.object.isRequired,
  accounts: PropTypes.arrayOf(PropTypes.object),
  getData: PropTypes.func.isRequired,
  setDates: PropTypes.func.isRequired,
  snoozeEvent: PropTypes.func.isRequired,
  unsnoozeEvent: PropTypes.func.isRequired,
  filtersActions: PropTypes.shape({
    get: PropTypes.func.isRequired,
    clear: PropTypes.func.isRequired,
    set: PropTypes.func.isRequired,
    clearSet: PropTypes.func.isRequired,
  }).isRequired,
  filters: PropTypes.shape({
    status: PropTypes.bool.isRequired,
    error: PropTypes.instanceOf(Error),
    values: PropTypes.arrayOf(
      PropTypes.shape({
        name: PropTypes.string.isRequired,
        desc: PropTypes.string.isRequired,
        rule: PropTypes.string.isRequired,
        data: PropTypes.isRequired,
        disabled: PropTypes.bool.isRequired
      })
    )
  }),
  setFilters: PropTypes.shape({
    status: PropTypes.bool.isRequired,
    error: PropTypes.instanceOf(Error),
    values: PropTypes.array
  }),
};

/* istanbul ignore next */
const mapStateToProps = ({aws, events}) => ({
  dates: events.dates,
  accounts: aws.accounts.selection,
  values: events.values,
  filters: events.getFilters,
  setFilters: events.setFilters
});

/* istanbul ignore next */
const mapDispatchToProps = (dispatch) => ({
  getData: (begin, end) => {
    dispatch(Actions.Events.getData(begin, end));
  },
  setDates: (startDate, endDate) => {
    dispatch(Actions.Events.setDates(startDate, endDate));
  },
  snoozeEvent: (id) => {
    dispatch(Actions.Events.snoozeEvent(id));
  },
  unsnoozeEvent: (id) => {
    dispatch(Actions.Events.unsnoozeEvent(id));
  },
  filtersActions: {
    get: () => {
      dispatch(Actions.Events.getFilters());
    },
    clear: () => {
      dispatch(Actions.Events.clearGetFilters());
    },
    set: (filters) => {
      dispatch(Actions.Events.setFilters(filters));
    },
    clearSet: () => {
      dispatch(Actions.Events.clearSetFilters());
    },
  }
});

export default connect(mapStateToProps, mapDispatchToProps)(EventsContainer);

