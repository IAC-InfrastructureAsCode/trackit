import ValuesReducer from '../valuesReducer';
import Constants from '../../../../constants';
import FilterReducer from "../filterReducer";

describe("ValuesReducer", () => {

  const id = "id";
  const costs = "costs";
  let state = {};
  state[id] = costs;

  it("handles initial state", () => {
    expect(ValuesReducer(undefined, {})).toEqual({});
  });

  it("handles set values state", () => {
    expect(ValuesReducer({}, { type: Constants.AWS_GET_COSTS_SUCCESS, id, costs })).toEqual(state);
  });

  it("handles error with values state", () => {
    expect(ValuesReducer(state, { type: Constants.AWS_GET_COSTS_ERROR })).toEqual({});
  });

  it("handles chart deletion state", () => {
    expect(ValuesReducer(state, { type: Constants.AWS_REMOVE_CHART, id })).toEqual({});
    expect(ValuesReducer(state, { type: Constants.AWS_REMOVE_CHART, id: "fakeID" })).toEqual(state);
  });

  it("handles wrong type state", () => {
    expect(ValuesReducer(state, { type: "" })).toEqual(state);
  });

});
