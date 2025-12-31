/// 앱 설정 모델
class AppConfig {
  final String connectionMode; // 'agent' or 'direct'
  final String agentServerURL;
  final int timeoutSeconds;

  AppConfig({
    this.connectionMode = 'direct',
    this.agentServerURL = 'http://172.24.10.6:8080',
    this.timeoutSeconds = 10,
  });

  factory AppConfig.fromJson(Map<String, dynamic> json) {
    return AppConfig(
      connectionMode: json['connectionMode'] as String? ?? 'direct',
      agentServerURL:
          json['agentServerURL'] as String? ?? 'http://172.24.10.6:8080',
      timeoutSeconds: json['timeoutSeconds'] as int? ?? 10,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'connectionMode': connectionMode,
      'agentServerURL': agentServerURL,
      'timeoutSeconds': timeoutSeconds,
    };
  }

  AppConfig copyWith({
    String? connectionMode,
    String? agentServerURL,
    int? timeoutSeconds,
  }) {
    return AppConfig(
      connectionMode: connectionMode ?? this.connectionMode,
      agentServerURL: agentServerURL ?? this.agentServerURL,
      timeoutSeconds: timeoutSeconds ?? this.timeoutSeconds,
    );
  }
}
