import { put, all } from 'redux-saga/effects';
import cleanUserTokenSaga from '../cleanUserTokenSaga';
import Constants from '../../../constants';

describe("Clean User Token Saga", () => {

  it("handless saga", () => {

    let saga = cleanUserTokenSaga();

    expect(saga.next().value)
      .toEqual(put({ type: Constants.CLEAN_USER_TOKEN_SUCCESS}));

    expect(saga.next().done).toBe(true);

  });


});
