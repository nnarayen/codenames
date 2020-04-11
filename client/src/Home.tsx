import * as React from 'react';
import axios from 'axios';

import { RouteComponentProps } from 'react-router-dom';

import Button from 'react-bootstrap/Button';
import FormControl from 'react-bootstrap/FormControl';
import InputGroup from 'react-bootstrap/InputGroup';

import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';

import { toast } from 'react-toastify';

import './stylesheets/home.scss';

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

  onKeyUp = (event: React.KeyboardEvent) => {
    if (event.key === 'Enter') {
      this.onJoinClick();
    }
  };

  onJoinClick = async () => {
    const identifier = this.textInput.current!.value;

    if (!identifier) {
      toast.error('Need to enter a game to join!');
      return;
    }

    try {
      await axios.get(`/api/games/${identifier}`); // check if game exists
      const { history } = this.props;
      history.push(`/games/${this.textInput.current!.value}`);
    } catch (error) {
      if (error.response.status === 404) { // game doesn't exist
        toast.error(`Game ${identifier} doesn't exist`);
      }
    }
  };

  onCreateClick = async () => {
    const result = (await axios.post('/api/games')).data as CreateResponse;

    const { history } = this.props;
    history.push(`/games/${result.identifier}`);
  };

  public render() {
    return (
      <div>
        <Container>
          <Row className="justify-content-center">
            <Col xs="4">
              <InputGroup>
                <FormControl ref={this.textInput} onKeyUp={this.onKeyUp} placeholder="Game identifier" aria-label="Game identifier" />
                <InputGroup.Append>
                  <Button ref={this.buttonInput} onClick={this.onJoinClick}>Join</Button>
                </InputGroup.Append>
              </InputGroup>
            </Col>
          </Row>
          <Row className="justify-content-center">
            <Button className="create-button" onClick={this.onCreateClick}>Create game</Button>
          </Row>
        </Container>
      </div>
    );
  }
}

export default Home;
