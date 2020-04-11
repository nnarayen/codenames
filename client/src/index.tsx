import React from 'react';
import ReactDOM from 'react-dom';
import { Route, Switch, BrowserRouter as Router } from 'react-router-dom';

import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';

import { ToastContainer } from 'react-toastify';
import Home from './Home';
import Gameboard from './Gameboard';

import './stylesheets/app.scss';
import './stylesheets/index.scss';

ReactDOM.render(
  <React.StrictMode>
    <Container>
      <Row className="justify-content-center">
        <h1 className="hero-title">
          <a href="/">
            Codenames
          </a>
        </h1>
      </Row>
    </Container>
    <Router>
      <Switch>
        <Route exact path="/" component={Home} />
        <Route exact path="/games/:identifier" component={Gameboard} />
      </Switch>
    </Router>
    <ToastContainer />
  </React.StrictMode>,
  document.getElementById('root'),
);
