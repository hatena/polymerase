// package: admin
// file: server/admin/admin.proto

import * as server_admin_admin_pb from "../../server/admin/admin_pb";
import * as github_com_taku_k_polymerase_pkg_status_statuspb_status_pb from "../../github.com/taku-k/polymerase/pkg/status/statuspb/status_pb";
export class Admin {
  static serviceName = "admin.Admin";
}
export namespace Admin {
  export class Backups {
    static readonly methodName = "Backups";
    static readonly service = Admin;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = server_admin_admin_pb.BackupsRequest;
    static readonly responseType = server_admin_admin_pb.BackupsResponse;
  }
}
