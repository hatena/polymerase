import { combineReducers, Reducer } from 'redux';
import backups from './backups';

export interface AdminUIState {
  backups: BackupStoreState;
}

export default combineReducers<AdminUIState>({
  backups
});


