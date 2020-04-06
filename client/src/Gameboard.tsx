import * as React from 'react';
import axios from 'axios';

import { RouteComponentProps } from 'react-router-dom';

import { MouseEvent } from 'react';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';

import Button from 'react-bootstrap/Button';
import ToggleButtonGroup from 'react-bootstrap/ToggleButtonGroup';
import ToggleButton from 'react-bootstrap/ToggleButton';

import Gameword from './Gameword';

const GRID_SIZE = 5;

interface GameboardParams {
  identifier: string;
}

interface GameboardProps extends RouteComponentProps<GameboardParams> { }

interface GameboardState {
  words: GameWord[];
  playerType: number;
  startingColor?: number;
  notFound?: boolean;
}

interface GameWord {
  word: string;
  color: number;
  revealed: boolean;
}

class Gameboard extends React.Component<GameboardProps, GameboardState> {
  constructor(props: GameboardProps) {
    super(props);
    this.state = {
      words: [],
      playerType: 0,
    };
  }

  onSocketMessage = (event: any) => {
    this.setState(JSON.parse(event.data) as GameboardState);
  };

  // register websocket, fetch game state
  componentDidMount = async () => {
    try {
      const ws = new WebSocket(`ws://localhost:8080/games/${this.props.match.params.identifier}/socket`);
      ws.onmessage = this.onSocketMessage;

      const result = await axios.get(`http://localhost:8080/games/${this.props.match.params.identifier}`);
      this.setState(result.data as GameboardState);
    } catch (error) {
      if (error.response.status === 404) {
        this.setState({
          notFound: true,
        });
      }
    }
  };

  // send click event to backend server
  _onWordClick = async (index: number) => {
    await axios.post(`http://localhost:8080/games/${this.props.match.params.identifier}/update`, {
      clicked: index,
    });
  };

  // split array into chucks of GRID_SIZE
  chunkWords(words: GameWord[]): GameWord[][] {
    const chunkedArr: GameWord[][] = [];
    for (let i = 0; i < words.length; i += GRID_SIZE) {
      chunkedArr.push(words.slice(i, i + GRID_SIZE));
    }

    return chunkedArr;
  }

  _onPlayerClick = (event: MouseEvent<HTMLButtonElement>) => {
    const value = (event.target as HTMLButtonElement).getAttribute('value');
    this.setState({
      playerType: Number(value),
    });
  };

  public render() {
    return (
      <Container>
        {
          this.state.notFound ? (
            <p>
              Game
              {this.props.match.params.identifier}
              {' '}
              does not exist idiot
            </p>
          ) : (
            this.chunkWords(this.state.words).map((chunk, index) => (
              <Row key={index}>
                {
                    chunk.map((info, chunkIndex) => {
                      const wordIndex = index * GRID_SIZE + chunkIndex;
                      return (
                        <Gameword
                          key={wordIndex}
                          word={info.word}
                          color={info.color}
                          revealed={info.revealed}
                          index={wordIndex}
                          onClick={this._onWordClick}
                          playerType={this.state.playerType}
                        />
                      );
                    })
                  }
              </Row>
            ))
          )
        }
        <br />
        <Row>
          <Col xs="2">
            <Button href="/">Home</Button>
          </Col>
          <Col xs="8">
            <ToggleButtonGroup type="radio" name="Player type" defaultValue={0}>
              <ToggleButton onClick={this._onPlayerClick} value={0}>Player</ToggleButton>
              <ToggleButton onClick={this._onPlayerClick} value={1}>Spymaster</ToggleButton>
            </ToggleButtonGroup>
          </Col>
        </Row>
      </Container>
    );
  }
}

export default Gameboard;
