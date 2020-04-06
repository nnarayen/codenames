import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import { Route, Switch, BrowserRouter as Router } from 'react-router-dom';
import Home from './Home';
import Gameboard from './Gameboard';

ReactDOM.render(
  <React.StrictMode>
    <Router>
      <Switch>
        <Route exact path="/" component={Home} />
        <Route exact path="/games/:identifier" component={Gameboard} />
      </Switch>
    </Router>
  </React.StrictMode>,
  document.getElementById('root'),
);
