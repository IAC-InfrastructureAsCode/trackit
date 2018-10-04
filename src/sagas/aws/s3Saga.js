import { put, call } from 'redux-saga/effects';
import API from '../../api';
import Constants from '../../constants';
import Moment from 'moment';
import {getAWSAccounts, getToken, getS3Dates} from "../misc";
import {setS3Dates, getS3Dates as getS3DatesLS, unsetS3Dates} from "../../common/localStorage";

export function* getS3DataSaga({ begin, end }) {
  try {
    const token = yield getToken();
    const accounts = yield getAWSAccounts();
    const res = yield call(API.AWS.S3.getData, token, begin, end, accounts);
    if (res.success === null) {
      yield put({type: Constants.LOGOUT_REQUEST});
      return;
    }
    if (res.success && res.hasOwnProperty("data") && !res.data.hasOwnProperty("error"))
      yield put({ type: Constants.AWS_GET_S3_DATA_SUCCESS, data: res.data });
    else
      throw Error("Error with request");
  } catch (error) {
    yield put({ type: Constants.AWS_GET_S3_DATA_ERROR, error });
  }
}

export function* saveS3DatesSaga() {
  const data = yield getS3Dates();
  setS3Dates(data);
}

export function* loadS3DatesSaga() {
  try {
    const data = yield call(getS3DatesLS);
    let dates;
    if (!data)
      dates = {
        startDate: Moment().subtract(1, "months").startOf("months"),
        endDate: Moment().subtract(1, "months").endOf("months")
      };
    else if (!Array.isArray(data))
      dates = {
        startDate: Moment(data.startDate),
        endDate: Moment(data.endDate)
      };
    else
      throw Error("Invalid data for S3 Analytics dates");
    yield put({type: Constants.AWS_INSERT_S3_DATES, dates});
    setS3Dates(dates);
  } catch (error) {
    yield put({ type: Constants.AWS_LOAD_S3_DATES_ERROR, error });
  }
}

export function* cleanS3DatesSaga() {
  yield call(unsetS3Dates);
}

