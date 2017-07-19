import * as React from 'react';
import * as style from './style.css';
import { TodoTextInput } from '../TodoTextInput';
import { AppBar, MenuItem } from 'material-ui';
import IconButton from 'material-ui/IconButton';
import BackupIcon from 'material-ui/svg-icons/action/backup';
import Drawer from 'material-ui/Drawer';
import ContentClear from 'material-ui/svg-icons/content/clear';
import Divider from 'material-ui/Divider';
import { Link } from 'react-router-dom';

export namespace Header {
  export interface Props {

  }

  export interface State {
    open: boolean
  }
}

export class Header extends React.Component<Header.Props, Header.State> {

  constructor(props?: Header.Props, context?: any) {
    super(props, context);
    this.state = {
      open: false,
    };
    this.handleLeftIconToggle = this.handleLeftIconToggle.bind(this);
    this.handleDrawerClose = this.handleDrawerClose.bind(this);
  }

  handleLeftIconToggle() {
    this.setState({open: !this.state.open});
  }

  handleDrawerClose() {
    this.setState({open: false})
  }

  render() {
    return (
      <div>
        <AppBar
          title="Polymerase UI"
          iconElementLeft={<IconButton><BackupIcon /></IconButton>}
          className={style.navbar}
          onLeftIconButtonTouchTap={this.handleLeftIconToggle}
        />
        <Drawer open={this.state.open}>
          <ContentClear onClick={this.handleDrawerClose}/>
          <Divider/>
          <MenuItem onTouchTap={this.handleDrawerClose}><Link to="/">Home</Link></MenuItem>
        </Drawer>
      </div>
    );
  }
}
