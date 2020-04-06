import * as React from 'react';
import Col from 'react-bootstrap/Col';

import './Gameword.css';

interface GamewordProps {
  word: string;
  color: number;
  revealed: boolean;
  index: number;
  onClick: Function;
}

class Gameword extends React.Component<GamewordProps, {}> {

  _getClassName() {
    return (this.props.revealed ? 'revealed' : '') +
      ` color-${this.props.color}` +
      ` gameWord`;
  }

  _onClick = () => {
    this.props.onClick(this.props.index);
  }

  public render() {
    return (
      <Col xs="2">
        <div
          onClick={this._onClick}
          className={this._getClassName()}>{this.props.word}
        </div>
      </Col>
    );
  }
}

export default Gameword;
