
import {Action, AnyAction, createStore} from 'redux';

const TRANSMIT = 'transmit';

function reducer(state = 1, action: AnyAction) {
  switch (action.type) {
    case TRANSMIT:
      return action.data;
    default:
      return state;
  }
}

export const store = createStore(reducer);

export const transmit = (data: any) => {
  return {type: TRANSMIT, data: data};
};
