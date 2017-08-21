// package: storagepb
// file: storage/storagepb/storage.proto

import * as jspb from "google-protobuf";

export class GetLatestToLSNRequest extends jspb.Message {
  getDb(): string;
  setDb(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetLatestToLSNRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetLatestToLSNRequest): GetLatestToLSNRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetLatestToLSNRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetLatestToLSNRequest;
  static deserializeBinaryFromReader(message: GetLatestToLSNRequest, reader: jspb.BinaryReader): GetLatestToLSNRequest;
}

export namespace GetLatestToLSNRequest {
  export type AsObject = {
    db: string,
  }
}

export class GetLatestToLSNResponse extends jspb.Message {
  getLsn(): string;
  setLsn(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetLatestToLSNResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetLatestToLSNResponse): GetLatestToLSNResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetLatestToLSNResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetLatestToLSNResponse;
  static deserializeBinaryFromReader(message: GetLatestToLSNResponse, reader: jspb.BinaryReader): GetLatestToLSNResponse;
}

export namespace GetLatestToLSNResponse {
  export type AsObject = {
    lsn: string,
  }
}

export class GetKeysAtPointRequest extends jspb.Message {
  getDb(): string;
  setDb(value: string): void;

  getFrom(): string;
  setFrom(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetKeysAtPointRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetKeysAtPointRequest): GetKeysAtPointRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetKeysAtPointRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetKeysAtPointRequest;
  static deserializeBinaryFromReader(message: GetKeysAtPointRequest, reader: jspb.BinaryReader): GetKeysAtPointRequest;
}

export namespace GetKeysAtPointRequest {
  export type AsObject = {
    db: string,
    from: string,
  }
}

export class GetKeysAtPointResponse extends jspb.Message {
  clearKeysList(): void;
  getKeysList(): Array<BackupFileInfo>;
  setKeysList(value: Array<BackupFileInfo>): void;
  addKeys(value?: BackupFileInfo, index?: number): BackupFileInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetKeysAtPointResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetKeysAtPointResponse): GetKeysAtPointResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetKeysAtPointResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetKeysAtPointResponse;
  static deserializeBinaryFromReader(message: GetKeysAtPointResponse, reader: jspb.BinaryReader): GetKeysAtPointResponse;
}

export namespace GetKeysAtPointResponse {
  export type AsObject = {
    keysList: Array<BackupFileInfo.AsObject>,
  }
}

export class BackupFileInfo extends jspb.Message {
  getStorageType(): string;
  setStorageType(value: string): void;

  getBackupType(): string;
  setBackupType(value: string): void;

  getKey(): string;
  setKey(value: string): void;

  getSize(): number;
  setSize(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BackupFileInfo.AsObject;
  static toObject(includeInstance: boolean, msg: BackupFileInfo): BackupFileInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: BackupFileInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BackupFileInfo;
  static deserializeBinaryFromReader(message: BackupFileInfo, reader: jspb.BinaryReader): BackupFileInfo;
}

export namespace BackupFileInfo {
  export type AsObject = {
    storageType: string,
    backupType: string,
    key: string,
    size: number,
  }
}

export class GetFileByKeyRequest extends jspb.Message {
  getKey(): string;
  setKey(value: string): void;

  getStorageType(): string;
  setStorageType(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetFileByKeyRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetFileByKeyRequest): GetFileByKeyRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetFileByKeyRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetFileByKeyRequest;
  static deserializeBinaryFromReader(message: GetFileByKeyRequest, reader: jspb.BinaryReader): GetFileByKeyRequest;
}

export namespace GetFileByKeyRequest {
  export type AsObject = {
    key: string,
    storageType: string,
  }
}

export class FileStream extends jspb.Message {
  getContent(): Uint8Array | string;
  getContent_asU8(): Uint8Array;
  getContent_asB64(): string;
  setContent(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FileStream.AsObject;
  static toObject(includeInstance: boolean, msg: FileStream): FileStream.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: FileStream, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FileStream;
  static deserializeBinaryFromReader(message: FileStream, reader: jspb.BinaryReader): FileStream;
}

export namespace FileStream {
  export type AsObject = {
    content: Uint8Array | string,
  }
}

export class PurgePrevBackupRequest extends jspb.Message {
  getDb(): string;
  setDb(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PurgePrevBackupRequest.AsObject;
  static toObject(includeInstance: boolean, msg: PurgePrevBackupRequest): PurgePrevBackupRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PurgePrevBackupRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PurgePrevBackupRequest;
  static deserializeBinaryFromReader(message: PurgePrevBackupRequest, reader: jspb.BinaryReader): PurgePrevBackupRequest;
}

export namespace PurgePrevBackupRequest {
  export type AsObject = {
    db: string,
  }
}

export class PurgePrevBackupResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PurgePrevBackupResponse.AsObject;
  static toObject(includeInstance: boolean, msg: PurgePrevBackupResponse): PurgePrevBackupResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PurgePrevBackupResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PurgePrevBackupResponse;
  static deserializeBinaryFromReader(message: PurgePrevBackupResponse, reader: jspb.BinaryReader): PurgePrevBackupResponse;
}

export namespace PurgePrevBackupResponse {
  export type AsObject = {
    message: string,
  }
}

export class GetBestStartTimeRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetBestStartTimeRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetBestStartTimeRequest): GetBestStartTimeRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetBestStartTimeRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetBestStartTimeRequest;
  static deserializeBinaryFromReader(message: GetBestStartTimeRequest, reader: jspb.BinaryReader): GetBestStartTimeRequest;
}

export namespace GetBestStartTimeRequest {
  export type AsObject = {
  }
}

export class GetBestStartTimeResponse extends jspb.Message {
  getMinute(): number;
  setMinute(value: number): void;

  getHour(): number;
  setHour(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetBestStartTimeResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetBestStartTimeResponse): GetBestStartTimeResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetBestStartTimeResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetBestStartTimeResponse;
  static deserializeBinaryFromReader(message: GetBestStartTimeResponse, reader: jspb.BinaryReader): GetBestStartTimeResponse;
}

export namespace GetBestStartTimeResponse {
  export type AsObject = {
    minute: number,
    hour: number,
  }
}

