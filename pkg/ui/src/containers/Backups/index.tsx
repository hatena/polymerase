import * as React from 'react';
import { RouteComponentProps } from 'react-router';
import { withStyles, createStyleSheet } from 'material-ui/styles';
import Table, { TableBody, TableCell, TableHead, TableRow } from 'material-ui/Table';
import Paper from 'material-ui/Paper';

namespace BackupTable {
  export interface Props {
    classes: any
  }
}

const styleSheet = createStyleSheet('BasicTable', theme => ({
  paper: {
    width: '100%',
    marginTop: theme.spacing.unit * 3,
    overflowX: 'auto',
  },
}));

let id = 0;
function createData(name, calories, fat, carbs, protein) {
  id += 1;
  return { id, name, calories, fat, carbs, protein };
}
const data = [
  createData('Frozen yoghurt', 159, 6.0, 24, 4.0),
  createData('Ice cream sandwich', 237, 9.0, 37, 4.3),
  createData('Eclair', 262, 16.0, 24, 6.0),
  createData('Cupcake', 305, 3.7, 67, 4.3),
  createData('Gingerbread', 356, 16.0, 49, 3.9),
];

class BackupTable extends React.Component<BackupTable.Props, {}> {
  render() {
    const classes = this.props.classes;

    return (
      <Paper className={classes.paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Dessert (100g serving)</TableCell>
              <TableCell numeric>Calories</TableCell>
              <TableCell numeric>Fat (g)</TableCell>
              <TableCell numeric>Carbs (g)</TableCell>
              <TableCell numeric>Protein (g)</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data.map(n => {
              return (
                <TableRow key={n.id}>
                  <TableCell>
                    {n.name}
                  </TableCell>
                  <TableCell numeric>
                    {n.calories}
                  </TableCell>
                  <TableCell numeric>
                    {n.fat}
                  </TableCell>
                  <TableCell numeric>
                    {n.carbs}
                  </TableCell>
                  <TableCell numeric>
                    {n.protein}
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </Paper>
    );
  }
}

const styledBackupTable = withStyles(styleSheet)(BackupTable);

export {
  styledBackupTable as BackupTable
};
