import React, {Component} from 'react';
import PropTypes from 'prop-types';
import Moment from 'moment';
import IntervalSelector from './IntervalSelector';

class IntervalNavigator extends Component {

  constructor(props) {
    super(props);
    this.previousDate = this.previousDate.bind(this);
    this.nextDate = this.nextDate.bind(this);
    this.updateInterval = this.updateInterval.bind(this);
  }

  previousDate = (e) => {
    e.preventDefault();
    let start;
    let end;
    switch (this.props.interval) {
      case "year":
        start = this.props.startDate.subtract(1, 'years');
        end = this.props.endDate.subtract(1, 'years');
        break;
      case "month":
        start = this.props.startDate.subtract(1, 'months');
        end = this.props.endDate.subtract(1, 'months').endOf('months');
        break;
      case "week":
        start = this.props.startDate.subtract(1, 'weeks');
        end = this.props.endDate.subtract(1, 'weeks');
        break;
      case "day":
      default:
        start = this.props.startDate.subtract(1, 'days');
        end = start;
    }
    this.props.setDatesFunc(start, end);
  };

  nextDate = (e) => {
    e.preventDefault();
    let start;
    let end;
    switch (this.props.interval) {
      case "year":
        start = this.props.startDate.add(1, 'years');
        end = this.props.endDate.add(1, 'years');
        break;
      case "month":
        start = this.props.startDate.add(1, 'months');
        end = this.props.endDate.add(1, 'months').endOf('months');
        break;
      case "week":
        start = this.props.startDate.add(1, 'weeks');
        end = this.props.endDate.add(1, 'weeks');
        break;
      case "day":
      default:
        start = this.props.startDate.add(1, 'days');
        end = start;
    }
    this.props.setDatesFunc(start, end);
  };

  getDate() {
    switch (this.props.interval) {
      case "year":
        return this.props.startDate.format('Y');
      case "month":
        return this.props.startDate.format('MMM Y');
      case "week":
        return (
          <div className="inline-block">
            {this.props.startDate.format('MMM Do Y')}
            &nbsp;
            <i className="fa fa-long-arrow-right"/>
            &nbsp;
            {this.props.endDate.format('MMM Do Y')}
          </div>
        );
      case "day":
      default:
        return this.props.startDate.format('MMM Do Y');
    }
  }

  updateInterval(interval) {
    this.props.setIntervalFunc(interval);
    let start;
    let end;
    switch (interval) {
      case "year":
        start = Moment().startOf('year');
        end = Moment().endOf('year');
        break;
      case "month":
        start = Moment().subtract(1, 'month').startOf('month');
        end = Moment().subtract(1, 'month').endOf('month');
        break;
      case "week":
        start = Moment().subtract(1, 'month').endOf('month').startOf('week');
        end = Moment().subtract(1, 'month').endOf('month').endOf('week');
        break;
      case "day":
      default:
        start = Moment().subtract(1, 'month').endOf('month');
        end = start;
    }
    return this.props.setDatesFunc(start, end);
  }

  render() {
    return(
      <div className="inline-block">
        <div className="inline-block btn-group">
          <button className="btn btn-default" onClick={this.previousDate}>
            <i className="fa fa-caret-left"/>
          </button>
          <div className="btn btn-default no-click">
            <i className="fa fa-calendar"/>
            &nbsp;
            {this.getDate()}
          </div>
          <button className="btn btn-default" onClick={this.nextDate}>
            <i className="fa fa-caret-right"/>
          </button>
        </div>
        <IntervalSelector interval={this.props.interval} setInterval={this.updateInterval}/>
      </div>
    );
  }

}

IntervalNavigator.propTypes = {
  startDate: PropTypes.object.isRequired,
  endDate: PropTypes.object.isRequired,
  setDatesFunc: PropTypes.func.isRequired,
  interval: PropTypes.string,
  setIntervalFunc: PropTypes.func
};

export default IntervalNavigator;
