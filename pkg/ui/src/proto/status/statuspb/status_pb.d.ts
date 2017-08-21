// package: statuspb
// file: status/statuspb/status.proto

import * as jspb from "google-protobuf";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";

export class DiskInfo extends jspb.Message {
  getTotal(): number;
  setTotal(value: number): void;

  getAvail(): number;
  setAvail(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DiskInfo.AsObject;
  static toObject(includeInstance: boolean, msg: DiskInfo): DiskInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DiskInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DiskInfo;
  static deserializeBinaryFromReader(message: DiskInfo, reader: jspb.BinaryReader): DiskInfo;
}

export namespace DiskInfo {
  export type AsObject = {
    total: number,
    avail: number,
  }
}

export class NodeInfo extends jspb.Message {
  hasDiskinfo(): boolean;
  clearDiskinfo(): void;
  getDiskinfo(): DiskInfo | undefined;
  setDiskinfo(value?: DiskInfo): void;

  getAddr(): string;
  setAddr(value: string): void;

  getStoreDir(): string;
  setStoreDir(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): NodeInfo.AsObject;
  static toObject(includeInstance: boolean, msg: NodeInfo): NodeInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: NodeInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): NodeInfo;
  static deserializeBinaryFromReader(message: NodeInfo, reader: jspb.BinaryReader): NodeInfo;
}

export namespace NodeInfo {
  export type AsObject = {
    diskinfo?: DiskInfo.AsObject,
    addr: string,
    storeDir: string,
  }
}

export class NodeInfoMap extends jspb.Message {
  getNodesMap(): jspb.Map<string, NodeInfo>;
  clearNodesMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): NodeInfoMap.AsObject;
  static toObject(includeInstance: boolean, msg: NodeInfoMap): NodeInfoMap.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: NodeInfoMap, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): NodeInfoMap;
  static deserializeBinaryFromReader(message: NodeInfoMap, reader: jspb.BinaryReader): NodeInfoMap;
}

export namespace NodeInfoMap {
  export type AsObject = {
    nodesMap: Array<[string, NodeInfo.AsObject]>,
  }
}

export class BackupInfo extends jspb.Message {
  hasFullBackup(): boolean;
  clearFullBackup(): void;
  getFullBackup(): FullBackupInfo | undefined;
  setFullBackup(value?: FullBackupInfo): void;

  clearIncBackupsList(): void;
  getIncBackupsList(): Array<IncBackupInfo>;
  setIncBackupsList(value: Array<IncBackupInfo>): void;
  addIncBackups(value?: IncBackupInfo, index?: number): IncBackupInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BackupInfo.AsObject;
  static toObject(includeInstance: boolean, msg: BackupInfo): BackupInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: BackupInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BackupInfo;
  static deserializeBinaryFromReader(message: BackupInfo, reader: jspb.BinaryReader): BackupInfo;
}

export namespace BackupInfo {
  export type AsObject = {
    fullBackup?: FullBackupInfo.AsObject,
    incBackupsList: Array<IncBackupInfo.AsObject>,
  }
}

export class FullBackupInfo extends jspb.Message {
  hasStoredTime(): boolean;
  clearStoredTime(): void;
  getStoredTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setStoredTime(value?: google_protobuf_timestamp_pb.Timestamp): void;

  getNodeName(): string;
  setNodeName(value: string): void;

  getHost(): string;
  setHost(value: string): void;

  getStoredType(): StoredType;
  setStoredType(value: StoredType): void;

  hasEndTime(): boolean;
  clearEndTime(): void;
  getEndTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setEndTime(value?: google_protobuf_timestamp_pb.Timestamp): void;

  getFileSize(): number;
  setFileSize(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FullBackupInfo.AsObject;
  static toObject(includeInstance: boolean, msg: FullBackupInfo): FullBackupInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: FullBackupInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FullBackupInfo;
  static deserializeBinaryFromReader(message: FullBackupInfo, reader: jspb.BinaryReader): FullBackupInfo;
}

export namespace FullBackupInfo {
  export type AsObject = {
    storedTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    nodeName: string,
    host: string,
    storedType: StoredType,
    endTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    fileSize: number,
  }
}

export class IncBackupInfo extends jspb.Message {
  hasStoredTime(): boolean;
  clearStoredTime(): void;
  getStoredTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setStoredTime(value?: google_protobuf_timestamp_pb.Timestamp): void;

  getNodeName(): string;
  setNodeName(value: string): void;

  getHost(): string;
  setHost(value: string): void;

  getStoredType(): StoredType;
  setStoredType(value: StoredType): void;

  hasEndTime(): boolean;
  clearEndTime(): void;
  getEndTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setEndTime(value?: google_protobuf_timestamp_pb.Timestamp): void;

  getFileSize(): number;
  setFileSize(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): IncBackupInfo.AsObject;
  static toObject(includeInstance: boolean, msg: IncBackupInfo): IncBackupInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: IncBackupInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): IncBackupInfo;
  static deserializeBinaryFromReader(message: IncBackupInfo, reader: jspb.BinaryReader): IncBackupInfo;
}

export namespace IncBackupInfo {
  export type AsObject = {
    storedTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    nodeName: string,
    host: string,
    storedType: StoredType,
    endTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    fileSize: number,
  }
}

export class AllBackups extends jspb.Message {
  getDbToBackupsMap(): jspb.Map<string, BackupInfo>;
  clearDbToBackupsMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AllBackups.AsObject;
  static toObject(includeInstance: boolean, msg: AllBackups): AllBackups.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AllBackups, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AllBackups;
  static deserializeBinaryFromReader(message: AllBackups, reader: jspb.BinaryReader): AllBackups;
}

export namespace AllBackups {
  export type AsObject = {
    dbToBackupsMap: Array<[string, BackupInfo.AsObject]>,
  }
}

export enum StoredType {
  LOCAL = 0,
}

