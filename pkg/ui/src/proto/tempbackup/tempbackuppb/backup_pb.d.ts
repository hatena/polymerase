// package: tempbackuppb
// file: tempbackup/tempbackuppb/backup.proto

import * as jspb from "google-protobuf";

export class FullBackupContentStream extends jspb.Message {
  getDb(): string;
  setDb(value: string): void;

  getContent(): Uint8Array | string;
  getContent_asU8(): Uint8Array;
  getContent_asB64(): string;
  setContent(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FullBackupContentStream.AsObject;
  static toObject(includeInstance: boolean, msg: FullBackupContentStream): FullBackupContentStream.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: FullBackupContentStream, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FullBackupContentStream;
  static deserializeBinaryFromReader(message: FullBackupContentStream, reader: jspb.BinaryReader): FullBackupContentStream;
}

export namespace FullBackupContentStream {
  export type AsObject = {
    db: string,
    content: Uint8Array | string,
  }
}

export class IncBackupContentStream extends jspb.Message {
  getDb(): string;
  setDb(value: string): void;

  getLsn(): string;
  setLsn(value: string): void;

  getContent(): Uint8Array | string;
  getContent_asU8(): Uint8Array;
  getContent_asB64(): string;
  setContent(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): IncBackupContentStream.AsObject;
  static toObject(includeInstance: boolean, msg: IncBackupContentStream): IncBackupContentStream.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: IncBackupContentStream, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): IncBackupContentStream;
  static deserializeBinaryFromReader(message: IncBackupContentStream, reader: jspb.BinaryReader): IncBackupContentStream;
}

export namespace IncBackupContentStream {
  export type AsObject = {
    db: string,
    lsn: string,
    content: Uint8Array | string,
  }
}

export class BackupReply extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  getKey(): string;
  setKey(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BackupReply.AsObject;
  static toObject(includeInstance: boolean, msg: BackupReply): BackupReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: BackupReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BackupReply;
  static deserializeBinaryFromReader(message: BackupReply, reader: jspb.BinaryReader): BackupReply;
}

export namespace BackupReply {
  export type AsObject = {
    message: string,
    key: string,
  }
}

export class PostCheckpointsRequest extends jspb.Message {
  getKey(): string;
  setKey(value: string): void;

  getContent(): Uint8Array | string;
  getContent_asU8(): Uint8Array;
  getContent_asB64(): string;
  setContent(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PostCheckpointsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: PostCheckpointsRequest): PostCheckpointsRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PostCheckpointsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PostCheckpointsRequest;
  static deserializeBinaryFromReader(message: PostCheckpointsRequest, reader: jspb.BinaryReader): PostCheckpointsRequest;
}

export namespace PostCheckpointsRequest {
  export type AsObject = {
    key: string,
    content: Uint8Array | string,
  }
}

export class PostCheckpointsResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PostCheckpointsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: PostCheckpointsResponse): PostCheckpointsResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PostCheckpointsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PostCheckpointsResponse;
  static deserializeBinaryFromReader(message: PostCheckpointsResponse, reader: jspb.BinaryReader): PostCheckpointsResponse;
}

export namespace PostCheckpointsResponse {
  export type AsObject = {
    message: string,
  }
}

