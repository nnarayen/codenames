import * as React from 'react';
import Col from 'react-bootstrap/Col';

import './stylesheets/Gameword.scss';

interface GamewordProps {
  word: string;
  color: number;
  revealed: boolean;
  index: number;
  onClick: Function;
  playerType: number;
}

class Gameword extends React.Component<GamewordProps, {}> {
  onClick = () => {
    const { onClick, index } = this.props;
    onClick(index);
  };

  getClassName() {
    const { revealed, color, playerType } = this.props;
    return `${revealed ? 'revealed' : ''
    } color-${color}`
      + ` player-${playerType}`
      + ' gameWord';
  }

  public render() {
    const { word } = this.props;
    return (
      <Col xs="2">
        <div
          tabIndex={0}
          role="button"
          onClick={this.onClick}
          className={this.getClassName()}
        >
          {word}
        </div>
      </Col>
    );
  }
}

export default Gameword;
