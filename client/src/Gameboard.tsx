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

import { toast } from 'react-toastify';

import Gameword from './Gameword';

import './stylesheets/Gameboard.scss';

const GRID_SIZE = 5;

// enums for possible colors
enum Color {
  Red = "Red",
  Blue = "Blue",
};

interface GameboardParams {
  identifier: string;
}

interface GameboardProps extends RouteComponentProps<GameboardParams> { }

interface GameboardState {
  words: GameWord[];
  playerType: number;
  startingColor?: number;
}

interface GameWord {
  word: string;
  color: number;
  revealed: boolean;
}

class Gameboard extends React.Component<GameboardProps, GameboardState> {
  // split array into chucks of GRID_SIZE
  static chunkWords(words: GameWord[]): GameWord[][] {
    const chunkedArr: GameWord[][] = [];
    for (let i = 0; i < words.length; i += GRID_SIZE) {
      chunkedArr.push(words.slice(i, i + GRID_SIZE));
    }

    return chunkedArr;
  }

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
    const { match } = this.props;
    try {
      const ws = new WebSocket(`wss://${window.location.host}/api/games/${match.params.identifier}/socket`);
      ws.onmessage = this.onSocketMessage;

      const result = await axios.get(`/api/games/${match.params.identifier}`);
      this.setState(result.data as GameboardState);
    } catch (error) {
      if (error.response.status === 404) { // game doesn't exist
        toast.error(`Game ${match.params.identifier} does not exist!`);
        const { history } = this.props;
        history.push('/');
      }
    }
  };

  // send click event to backend server
  onWordClick = async (index: number) => {
    const { match } = this.props;
    await axios.post(`/api/games/${match.params.identifier}/update`, {
      clicked: index,
    });
  };

  onPlayerClick = (event: MouseEvent<HTMLButtonElement>) => {
    const value = (event.target as HTMLButtonElement).getAttribute('value');
    this.setState({
      playerType: Number(value),
    });
  };

  // calculate the number of remaining words
  calculateRemaining = () => {
    const remainingWords = {
      [Color.Red]: 0,
      [Color.Blue]: 0,
    };

    this.state.words.forEach((word) => {
      if (word.color === 0) {
        remainingWords[Color.Red] += (word.revealed ? 0 : 1);
      } else if (word.color === 1) {
        remainingWords[Color.Blue] += (word.revealed ? 0 : 1);
      }
    });

    return remainingWords;
  }

  public render() {
    const { words, playerType } = this.state;
    const remainingWords = this.calculateRemaining();
    return (
      <Container>
        {
          Gameboard.chunkWords(words).map((chunk, index) => (
            <Row key={index} className="justify-content-center game-row">
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
                        onClick={this.onWordClick}
                        playerType={playerType}
                      />
                    );
                  })
                }
            </Row>
          ))
        }
        <Row className="justify-content-center game-buttons">
          <Col xs="2">
            <Button href="/">Home</Button>
          </Col>
          <Col xs="2" className="remaining">
            <span className="remaining-0">
              Red:
              { ` ${remainingWords[Color.Red]}` }
            </span>
            |
            <span className="remaining-1">
              Blue:
              { ` ${remainingWords[Color.Blue]}` }
            </span>
          </Col>
          <Col xs="2">
            <ToggleButtonGroup type="radio" name="Player type" defaultValue={0}>
              <ToggleButton onClick={this.onPlayerClick} value={0}>Player</ToggleButton>
              <ToggleButton onClick={this.onPlayerClick} value={1}>Spymaster</ToggleButton>
            </ToggleButtonGroup>
          </Col>
        </Row>
      </Container>
    );
  }
}

export default Gameboard;
