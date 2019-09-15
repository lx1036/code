import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
// import App from './App';
import * as serviceWorker from './serviceWorker';
import { createStore, Action } from "redux";
import { Provider } from "react-redux";
import Counter from './Components/Counter';
import {INCREMENT, DECREMENT, RESET} from "./redux/actions";
import {Game} from './Demo/Game/Game';

////////////////////////////Counter Demo//////////////////////////////////////////////////


const initialState = { count: 0 };
const reducer = (state = initialState, action: Action) => {
  console.log("reducer", state, action);

  switch (action.type) {
    case INCREMENT:
      return {
        count: state.count + 1
      };
    case DECREMENT:
      return {
        count: state.count - 1
      };
    case RESET:
      return {
        count: state.count
      };
    default:
      return {
        count: state.count
      };
  }
};
const store = createStore(reducer, (window as any).__REDUX_DEVTOOLS_EXTENSION__ && (window as any).__REDUX_DEVTOOLS_EXTENSION__());
store.dispatch({ type: INCREMENT });
store.dispatch({ type: INCREMENT });
store.dispatch({ type: DECREMENT });
store.dispatch({ type: RESET });


const App = () => (
  <Provider store={store}>
    <Counter />
  </Provider>
);
////////////////////////////Counter Demo//////////////////////////////////////////////////



// ReactDOM.render(<App />, document.getElementById('root'));
ReactDOM.render(<Game />, document.getElementById('root'));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
