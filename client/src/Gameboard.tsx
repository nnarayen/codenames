import * as React from 'react';
import axios from 'axios';

import Container from 'react-bootstrap/Container'
import Row from 'react-bootstrap/Row'

import Gameword from './Gameword';

const GRID_SIZE = 5;

interface GameboardProps {
  identifier: string;
}

interface GameboardState {
  words: GameWord[];
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
    };
  }

  onSocketMessage = (event: any) => {
    this.setState(JSON.parse(event.data) as GameboardState);
  }

  // register websocket, fetch game state
  componentDidMount = async () => {
    try {
      const ws = new WebSocket(`ws://localhost:8080/games/${this.props.identifier}/socket`);
      ws.onmessage = this.onSocketMessage;

      const result = await axios.get(`http://localhost:8080/games/${this.props.identifier}`);
      this.setState(result.data as GameboardState);
    } catch (error) {
      if (error.response.status === 404) {
        this.setState({
          notFound: true,
        })
      }
    }
  }

  _onWordClick = async (index: number) => {
    await axios.post(`http://localhost:8080/games/${this.props.identifier}/update`, {
      clicked: index,
    });
  }

  // split array into chucks of GRID_SIZE
  chunkWords(words: GameWord[]): GameWord[][] {
    const chunkedArr: GameWord[][] = [];
    for (let i = 0; i < words.length; i += GRID_SIZE) {
      chunkedArr.push(words.slice(i, i + GRID_SIZE));
    }

    return chunkedArr;
  }

  public render() {
    return (
      <Container>
        {
          this.state.notFound ? (
            <p>Game {this.props.identifier} does not exist idiot</p>
          ) : (
            this.chunkWords(this.state.words).map((chunk, index) => {
              return (
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
                        />
                      );
                    })
                  }
                </Row>
              );
            })
          )
        }
      </Container>
    );
  }
}

export default Gameboard;
