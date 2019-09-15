import React from "react";
import { connect } from "react-redux";
import { INCREMENT, DECREMENT } from "../redux/actions";

class Counter extends React.Component<{dispatch: any, count: any}, any> {
  // state = { count: 0 };

  increment = () => {
    // this.setState({
    //   count: this.state.count + 1
    // });

    this.props.dispatch({ type: INCREMENT });
  };

  decrement = () => {
    // this.setState({
    //   count: this.state.count - 1
    // });
    this.props.dispatch({ type: DECREMENT });
  };

  render() {
    return (
      <div className="counter">
        <h2>Counter</h2>
        <div>
          <button onClick={this.decrement}>-</button>
          {/* <span className="count">{this.state.count}</span> */}
          <span className="count">{this.props.count}</span>
          <button onClick={this.increment}>+</button>
        </div>
      </div>
    );
  }
}

// export default Counter;
const mapStateToProps = (state: any) => ({ count: state.count });
export default connect(mapStateToProps)(Counter);
