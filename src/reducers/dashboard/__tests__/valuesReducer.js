import ValuesReducer from '../valuesReducer';
import Constants from '../../../constants';

describe("ValuesReducer", () => {

  const id = "id";
  const values = {
    value1: 1,
    value2: 2
  };

  let state = (data) => {
    let result = {};
    result[id] = data;
    return result;
  };

  const defaultValue = {};
  const requestedValue = state({status: false});
  const successValue = state({status: true, values});
  const errorValue = state({status: true, error: Error()});
  const cleared = {};

  it("handles initial state", () => {
    expect(ValuesReducer(undefined, {})).toEqual(defaultValue);
  });

  it("handles get values requested state", () => {
    expect(ValuesReducer(defaultValue, { type: Constants.DASHBOARD_GET_VALUES, id })).toEqual(requestedValue);
  });

  it("handles set values state", () => {
    expect(ValuesReducer(defaultValue, { type: Constants.DASHBOARD_GET_VALUES_SUCCESS, id, data: values })).toEqual(successValue);
  });

  it("handles error with values state", () => {
    expect(ValuesReducer(successValue, { type: Constants.DASHBOARD_GET_VALUES_ERROR, id, error: Error() })).toEqual(errorValue);
  });

  it("handles chart deletion state", () => {
    expect(ValuesReducer(successValue, { type: Constants.DASHBOARD_REMOVE_ITEM, id })).toEqual(cleared);
    expect(ValuesReducer(successValue, { type: Constants.DASHBOARD_REMOVE_ITEM, id: "fakeID" })).toEqual(successValue);
  });

  it("handles wrong type state", () => {
    expect(ValuesReducer(successValue, { type: "" })).toEqual(successValue);
  });

});
