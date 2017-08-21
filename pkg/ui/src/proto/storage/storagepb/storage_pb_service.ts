// package: storagepb
// file: storage/storagepb/storage.proto

import * as storage_storagepb_storage_pb from "../../storage/storagepb/storage_pb";
export class StorageService {
  static serviceName = "storagepb.StorageService";
}
export namespace StorageService {
  export class GetLatestToLSN {
    static readonly methodName = "GetLatestToLSN";
    static readonly service = StorageService;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = storage_storagepb_storage_pb.GetLatestToLSNRequest;
    static readonly responseType = storage_storagepb_storage_pb.GetLatestToLSNResponse;
  }
  export class GetKeysAtPoint {
    static readonly methodName = "GetKeysAtPoint";
    static readonly service = StorageService;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = storage_storagepb_storage_pb.GetKeysAtPointRequest;
    static readonly responseType = storage_storagepb_storage_pb.GetKeysAtPointResponse;
  }
  export class GetFileByKey {
    static readonly methodName = "GetFileByKey";
    static readonly service = StorageService;
    static readonly requestStream = false;
    static readonly responseStream = true;
    static readonly requestType = storage_storagepb_storage_pb.GetFileByKeyRequest;
    static readonly responseType = storage_storagepb_storage_pb.FileStream;
  }
  export class PurgePrevBackup {
    static readonly methodName = "PurgePrevBackup";
    static readonly service = StorageService;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = storage_storagepb_storage_pb.PurgePrevBackupRequest;
    static readonly responseType = storage_storagepb_storage_pb.PurgePrevBackupResponse;
  }
  export class GetBestStartTime {
    static readonly methodName = "GetBestStartTime";
    static readonly service = StorageService;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = storage_storagepb_storage_pb.GetBestStartTimeRequest;
    static readonly responseType = storage_storagepb_storage_pb.GetBestStartTimeResponse;
  }
}
