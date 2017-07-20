import * as React from 'react';
import * as style from './style.css';
import AppBar from 'material-ui/AppBar';
import Toolbar from 'material-ui/Toolbar';
import Typography from 'material-ui/Typography';
import IconButton from 'material-ui/IconButton';
import MenuIcon from 'material-ui-icons/Menu';
import List, { ListItem, ListItemIcon, ListItemText } from 'material-ui/List';
import Drawer from 'material-ui/Drawer';
import { Link } from 'react-router-dom';
import BackupIcon from 'material-ui-icons/Backup';

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

    this.handleDrawerOpen = this.handleDrawerOpen.bind(this);
    this.handleDrawerClose = this.handleDrawerClose.bind(this);
  }

  handleDrawerOpen() {
    this.setState({open: true})
  }

  handleDrawerClose() {
    this.setState({open: false})
  }

  render() {
    const sideList = (
      <div>
        <List className={style.list} disablePadding>
          <ListItem button>
            <ListItemIcon>
              <BackupIcon />
            </ListItemIcon>
          <ListItemText primary="Backups" />
        </ListItem>
        </List>
      </div>
    );

    return (
      <div>
        <AppBar position="static">
          <Toolbar>
            <IconButton color="contrast" aria-label="Menu">
              <MenuIcon onClick={this.handleDrawerOpen} />
            </IconButton>
            <Typography type="title" color="inherit" className="initial">
              Polymerase UI
            </Typography>
          </Toolbar>
        </AppBar>
        <Drawer
          open={this.state.open}
          onRequestClose={this.handleDrawerClose}
          onClick={this.handleDrawerClose}
        >
          {sideList}
        </Drawer>
      </div>
    );
  }
}
