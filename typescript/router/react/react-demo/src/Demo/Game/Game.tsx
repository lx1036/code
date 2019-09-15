import React from 'react';
import './game.css';

class Square extends React.Component<any, any> {
  constructor(props: any) {
    super(props);

    this.state = {
      value: null,
    };
  }

  render() {
    return (
      /*<button className="square" onClick={() => {alert('click');}}>*/
      <button className="square" onClick={() => this.setState({value: 'X'})}>
        {this.state.value}
      </button>
    );
  }
}

class Board extends React.Component<any, any> {

  constructor(props: any) {
    super(props);

    this.state = {
      squares: Array(9).fill(null),
    };
  }

  renderSquare(i: any) {
    return <Square value={this.state.squares[i]} />;
  }

  render() {
    const status = 'Next player: X';

    return (
      <div>
        <div className="status">{status}</div>
        <div className="board-row">
          {this.renderSquare(0)}
          {this.renderSquare(1)}
          {this.renderSquare(2)}
        </div>
        <div className="board-row">
          {this.renderSquare(3)}
          {this.renderSquare(4)}
          {this.renderSquare(5)}
        </div>
        <div className="board-row">
          {this.renderSquare(6)}
          {this.renderSquare(7)}
          {this.renderSquare(8)}
        </div>
      </div>
    );
  }
}

export class Game extends React.Component {
  render() {
    return (
      <div className="game">
        <div className="game-board">
          <Board />
        </div>
        <div className="game-info">
          <div>{/* status */}</div>
          <ol>{/* TODO */}</ol>
        </div>
      </div>
    );
  }
}


// ========================================

// ReactDOM.render(
//   <Game />,
//   document.getElementById('root')
// );
