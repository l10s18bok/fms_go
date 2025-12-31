/// 배포 결과 모델
class DeployResult {
  final String ip;
  final String status;
  final List<RuleResult> results;

  DeployResult({
    required this.ip,
    required this.status,
    required this.results,
  });

  factory DeployResult.fromJson(Map<String, dynamic> json) {
    return DeployResult(
      ip: json['ip'] as String? ?? '',
      status: json['status'] as String? ?? '',
      results: (json['results'] as List<dynamic>?)
              ?.map((e) => RuleResult.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'ip': ip,
      'status': status,
      'results': results.map((e) => e.toJson()).toList(),
    };
  }
}

/// 규칙별 결과 모델
class RuleResult {
  final String rule;
  final String text;
  final String status; // ok/error/unfind/validation
  final String reason;

  RuleResult({
    required this.rule,
    required this.text,
    required this.status,
    required this.reason,
  });

  factory RuleResult.fromJson(Map<String, dynamic> json) {
    return RuleResult(
      rule: json['rule'] as String? ?? '',
      text: json['text'] as String? ?? '',
      status: json['status'] as String? ?? '',
      reason: json['reason'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'rule': rule,
      'text': text,
      'status': status,
      'reason': reason,
    };
  }
}
