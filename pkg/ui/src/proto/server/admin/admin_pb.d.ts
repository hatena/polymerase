// package: admin
// file: server/admin/admin.proto

import * as jspb from "google-protobuf";
import * as github_com_taku_k_polymerase_pkg_status_statuspb_status_pb from "../../github.com/taku-k/polymerase/pkg/status/statuspb/status_pb";

export class BackupsRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BackupsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: BackupsRequest): BackupsRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: BackupsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BackupsRequest;
  static deserializeBinaryFromReader(message: BackupsRequest, reader: jspb.BinaryReader): BackupsRequest;
}

export namespace BackupsRequest {
  export type AsObject = {
  }
}

export class BackupsResponse extends jspb.Message {
  clearBackupsList(): void;
  getBackupsList(): Array<BackupInfoWithKey>;
  setBackupsList(value: Array<BackupInfoWithKey>): void;
  addBackups(value?: BackupInfoWithKey, index?: number): BackupInfoWithKey;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BackupsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: BackupsResponse): BackupsResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: BackupsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BackupsResponse;
  static deserializeBinaryFromReader(message: BackupsResponse, reader: jspb.BinaryReader): BackupsResponse;
}

export namespace BackupsResponse {
  export type AsObject = {
    backupsList: Array<BackupInfoWithKey.AsObject>,
  }
}

export class BackupInfoWithKey extends jspb.Message {
  getKey(): string;
  setKey(value: string): void;

  hasInfo(): boolean;
  clearInfo(): void;
  getInfo(): github_com_taku_k_polymerase_pkg_status_statuspb_status_pb.BackupInfo | undefined;
  setInfo(value?: github_com_taku_k_polymerase_pkg_status_statuspb_status_pb.BackupInfo): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BackupInfoWithKey.AsObject;
  static toObject(includeInstance: boolean, msg: BackupInfoWithKey): BackupInfoWithKey.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: BackupInfoWithKey, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BackupInfoWithKey;
  static deserializeBinaryFromReader(message: BackupInfoWithKey, reader: jspb.BinaryReader): BackupInfoWithKey;
}

export namespace BackupInfoWithKey {
  export type AsObject = {
    key: string,
    info?: github_com_taku_k_polymerase_pkg_status_statuspb_status_pb.BackupInfo.AsObject,
  }
}

