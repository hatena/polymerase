// package: tempbackuppb
// file: tempbackup/tempbackuppb/backup.proto

import * as tempbackup_tempbackuppb_backup_pb from "../../tempbackup/tempbackuppb/backup_pb";
export class BackupTransferService {
  static serviceName = "tempbackuppb.BackupTransferService";
}
export namespace BackupTransferService {
  export class TransferFullBackup {
    static readonly methodName = "TransferFullBackup";
    static readonly service = BackupTransferService;
    static readonly requestStream = true;
    static readonly responseStream = false;
    static readonly requestType = tempbackup_tempbackuppb_backup_pb.FullBackupContentStream;
    static readonly responseType = tempbackup_tempbackuppb_backup_pb.BackupReply;
  }
  export class TransferIncBackup {
    static readonly methodName = "TransferIncBackup";
    static readonly service = BackupTransferService;
    static readonly requestStream = true;
    static readonly responseStream = false;
    static readonly requestType = tempbackup_tempbackuppb_backup_pb.IncBackupContentStream;
    static readonly responseType = tempbackup_tempbackuppb_backup_pb.BackupReply;
  }
  export class PostCheckpoints {
    static readonly methodName = "PostCheckpoints";
    static readonly service = BackupTransferService;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = tempbackup_tempbackuppb_backup_pb.PostCheckpointsRequest;
    static readonly responseType = tempbackup_tempbackuppb_backup_pb.PostCheckpointsResponse;
  }
}
