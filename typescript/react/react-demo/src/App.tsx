import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';
import {BrowserRouter as Router} from 'react-router-dom';
import {Admin} from './Components/Admin';

class App extends Component {
  render() {
    return (
      <Router>
        <Admin>

        </Admin>
      </Router>
    );
  }
}

export default App;
