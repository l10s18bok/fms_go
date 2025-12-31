import 'deploy_result.dart';

/// 배포 이력 모델
class DeployHistory {
  final int id;
  final DateTime timestamp;
  final String deviceIp;
  final String templateVersion;
  final String status; // success/fail/error
  final List<RuleResult> results;

  DeployHistory({
    required this.id,
    required this.timestamp,
    required this.deviceIp,
    required this.templateVersion,
    required this.status,
    required this.results,
  });

  factory DeployHistory.fromJson(Map<String, dynamic> json) {
    return DeployHistory(
      id: json['id'] as int? ?? 0,
      timestamp: DateTime.tryParse(json['timestamp'] as String? ?? '') ??
          DateTime.now(),
      deviceIp: json['deviceIp'] as String? ?? '',
      templateVersion: json['templateVersion'] as String? ?? '',
      status: json['status'] as String? ?? '',
      results: (json['results'] as List<dynamic>?)
              ?.map((e) => RuleResult.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'timestamp': timestamp.toIso8601String(),
      'deviceIp': deviceIp,
      'templateVersion': templateVersion,
      'status': status,
      'results': results.map((e) => e.toJson()).toList(),
    };
  }
}
