import React from 'react';
import {connect} from "react-redux";
import {transmit} from "../redux/store";
import {NavLink} from "react-router-dom";
import {Input, Button, Menu, Avatar, Divider, Icon} from 'antd';
import PropTypes from "prop-types";

class Header extends React.Component<any, any> {
  state = {
    value: ''
  };

  static propTypes = {
    transmit: PropTypes.func.isRequired
  };

  getInputValue = (event: any) => {
    let value = event.target.value;
    this.setState({
      value
    });
  };

  menuSearch = () => {
    let value = this.state.value;
    this.props.transmit(value)
  };

  render() {
    return (
      <div className="container">
        <div className="topbar-container">
          <div className="logo">
            <NavLink to="/home"><span>小帮厨</span></NavLink>
          </div>
          <Input
            style={{width: '22%'}}
            placeholder="搜索菜谱、食材"
            onChange={event => this.getInputValue(event)}
            allowClear
            size="large"
            onClick={this.menuSearch}
          />
          <NavLink to="/search">
            <Button type="primary" icon="search" size="large" onClick={this.menuSearch}>搜菜谱</Button>
          </NavLink>
          <div className="topbar-menu">
            <Menu mode="horizontal">
              <Menu.SubMenu title={<span className="submenu-title-wrapper">菜谱分类</span>}>
                <Menu.ItemGroup title="常用主题" />
                <Menu.ItemGroup title="常见食材" />
                <Menu.ItemGroup title="时令食材" />
              </Menu.SubMenu>
              <Menu.Item key="alipay"><NavLink to="/topic">话题</NavLink></Menu.Item>
              <Menu.Item key="mail"><NavLink to="/menu">菜单</NavLink></Menu.Item>
              <Menu.Item key="app"><NavLink to="/collections">我的主页</NavLink></Menu.Item>
            </Menu>
          </div>

          <div className="avatar">
            <Avatar style={{ color: '#f56a00', backgroundColor: '#fde3cf' }}>U</Avatar>
            <Divider type="vertical"/>
            <NavLink to="/collections"><Icon type="book" style={{ fontSize: 25 }}/></NavLink>
          </div>
        </div>
      </div>
    );
  }
}

export default connect(state => ({keyword: state}), {transmit: transmit})(Header);
