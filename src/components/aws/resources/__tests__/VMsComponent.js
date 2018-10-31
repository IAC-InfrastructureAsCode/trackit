import React from 'react';
import { VMsComponent } from '../VMsComponent';
import { shallow } from 'enzyme';
import Spinner from "react-spinkit";
import ReactTable from 'react-table';
import Moment from 'moment';
import Misc from '../../../misc';

const Tooltip = Misc.Popover;

const props = {
  getData: jest.fn(),
  clear: jest.fn(),
  dates: {
    startDate: Moment().startOf("months"),
    endDate: Moment().endOf("months")
  }
};

const propsWithData = {
  ...props,
  data: {
    status: true,
    value: [
      {
        account: '420',
        reportDate: Moment().toISOString(),
        instance: {
          id: '42',
          state: 'running',
          region: 'us-west-1',
          keyPair: 'key',
          type: 'type',
          purchasing: 'value',
          tags: {
            Name: 'name'
          },
          costs: {
            instance: 42
          },
          stats: {
            cpu: {
              average: 42,
              peak: 42
            },
            network: {
              in: 42,
              out: 42
            },
            volumes: {
              read: {
                internal: 42
              },
              write: {
                internal: 42
              }
            }
          }
        }
      }
    ]
  }
};

const propsLoading = {
  ...props,
  data: {
    status: false,
    value: null
  }
};

const propsWithError = {
  ...props,
  data: {
    status: true,
    error: Error()
  }
};

describe('<VMsComponent />', () => {

  it('renders a <VMsComponent /> component', () => {
    const wrapper = shallow(<VMsComponent {...propsWithData}/>);
    expect(wrapper.length).toBe(1);
  });

  it('renders a <Tooltip /> component', () => {
    const wrapper = shallow(<VMsComponent {...propsWithData}/>);
    const tooltip = wrapper.find(Tooltip);
    expect(tooltip.length).toBe(1);
  });

  it('renders a <ReactTable /> component', () => {
    const wrapper = shallow(<VMsComponent {...propsWithData}/>);
    const table = wrapper.find(ReactTable);
    expect(table.length).toBe(1);
  });

  it('renders a <Spinner /> component when data is loading', () => {
    const wrapper = shallow(<VMsComponent {...propsLoading}/>);
    const spinner = wrapper.find(Spinner);
    expect(spinner.length).toBe(1);
  });

  it('renders an <div class="alert" /> component when data is not available', () => {
    const wrapper = shallow(<VMsComponent {...propsWithError}/>);
    const alert = wrapper.find("div.alert");
    expect(alert.length).toBe(1);
  });

});
