import * as React from 'react';
import axios from 'axios';

import { RouteComponentProps } from 'react-router-dom';

import Button from 'react-bootstrap/Button';
import FormControl from 'react-bootstrap/FormControl';
import InputGroup from 'react-bootstrap/InputGroup';

import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';

interface HomeProps extends RouteComponentProps<any> { }

interface CreateResponse {
  identifier: string;
}

class Home extends React.Component<HomeProps, {}> {
  private readonly textInput: React.RefObject<HTMLInputElement & FormControl>;

  private readonly buttonInput: React.RefObject<Button & HTMLButtonElement>;

  constructor(props: any) {
    super(props);
    this.textInput = React.createRef();
    this.buttonInput = React.createRef();
  }

  _onKeyUp = (event: React.KeyboardEvent) => {
    if (event.key === 'Enter') {
      this._onJoinClick();
    }
  };

  _onJoinClick = () => {
    this.props.history.push(`/games/${this.textInput.current!.value}`);
  };

  _onCreateClick = async () => {
    const result = (await axios.post('http://localhost:8080/games')).data as CreateResponse;
    this.props.history.push(`/games/${result.identifier}`);
  };

  public render() {
    return (
      <div id="home-container">
        <Container>
          <Row>
            <h2>Codenames</h2>
          </Row>
          <Row>
            <Col xs="4">
              <InputGroup>
                <FormControl ref={this.textInput} onKeyUp={this._onKeyUp} placeholder="Game identifier" aria-label="Game identifier" />
                <InputGroup.Append>
                  <Button ref={this.buttonInput} onClick={this._onJoinClick}>Join</Button>
                </InputGroup.Append>
              </InputGroup>
            </Col>
          </Row>
          <br />
          <Row>
            <Button onClick={this._onCreateClick}>Create game</Button>
          </Row>
        </Container>
      </div>
    );
  }
}

export default Home;
