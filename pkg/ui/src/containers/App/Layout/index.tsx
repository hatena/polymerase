import * as React from 'react';
import * as style from '../style.css';
import { Header } from '../../../components';
import { RouteComponentProps } from 'react-router';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';

export class Layout extends React.Component<RouteComponentProps<void>, {}> {
  render() {
    const { children } = this.props;
    return (
      <MuiThemeProvider>
        <div className={style.normal}>
          <Header />
          {children}
        </div>
      </MuiThemeProvider>
    )
  }
}
