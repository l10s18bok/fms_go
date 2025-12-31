import 'deploy_result.dart';

/// 방화벽 장비 모델
class Firewall {
  final int index;
  final String deviceName;
  final String serverStatus;
  final String deployStatus;
  final String version;
  final DeployResult? deployResult;

  Firewall({
    required this.index,
    required this.deviceName,
    this.serverStatus = '-',
    this.deployStatus = '-',
    this.version = '-',
    this.deployResult,
  });

  factory Firewall.fromJson(Map<String, dynamic> json) {
    return Firewall(
      index: json['index'] as int? ?? -1,
      deviceName: json['deviceName'] as String? ?? '',
      serverStatus: json['serverStatus'] as String? ?? '-',
      deployStatus: json['deployStatus'] as String? ?? '-',
      version: json['version'] as String? ?? '-',
      deployResult: json['deployResult'] != null
          ? DeployResult.fromJson(json['deployResult'] as Map<String, dynamic>)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'index': index,
      'deviceName': deviceName,
      'serverStatus': serverStatus,
      'deployStatus': deployStatus,
      'version': version,
      'deployResult': deployResult?.toJson(),
    };
  }

  Firewall copyWith({
    int? index,
    String? deviceName,
    String? serverStatus,
    String? deployStatus,
    String? version,
    DeployResult? deployResult,
  }) {
    return Firewall(
      index: index ?? this.index,
      deviceName: deviceName ?? this.deviceName,
      serverStatus: serverStatus ?? this.serverStatus,
      deployStatus: deployStatus ?? this.deployStatus,
      version: version ?? this.version,
      deployResult: deployResult ?? this.deployResult,
    );
  }
}
