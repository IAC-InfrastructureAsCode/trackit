import React from 'react';
import { Navigation } from '../Navigation';
import { shallow } from "enzyme/build/index";

const props = {
  signOut: jest.fn()
};

describe('<Navigation />', () => {

  it('renders a <Navigation /> component', () => {
    const wrapper = shallow(<Navigation {...props}/>);
    expect(wrapper.length).toEqual(1);
  });

  it('dispatches a logout action', () => {
    const wrapper = shallow(<Navigation {...props}/>);
    wrapper.setState({userMenuExpanded: true});
    const logout = wrapper.find('a');
    expect(props.signOut.mock.calls.length).toBe(0);
    logout.prop('onClick')();
    expect(props.signOut.mock.calls.length).toBe(1);
  });

  it('renders without user menu', () => {
    const wrapper = shallow(<Navigation {...props}/>);
    expect(wrapper.state('userMenuExpanded')).toBe(false);
  });

  it('can expand user menu', () => {
    const wrapper = shallow(<Navigation {...props}/>);
    expect(wrapper.state('userMenuExpanded')).toBe(false);
    wrapper.find('button').prop('onClick')();
    expect(wrapper.state('userMenuExpanded')).toBe(true);
  });

});
