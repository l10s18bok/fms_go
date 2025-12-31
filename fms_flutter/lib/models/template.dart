/// 방화벽 규칙 템플릿 모델
class Template {
  final String version;
  final String contents;

  Template({
    required this.version,
    required this.contents,
  });

  factory Template.fromJson(Map<String, dynamic> json) {
    return Template(
      version: json['version'] as String? ?? '',
      contents: json['contents'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'version': version,
      'contents': contents,
    };
  }

  Template copyWith({
    String? version,
    String? contents,
  }) {
    return Template(
      version: version ?? this.version,
      contents: contents ?? this.contents,
    );
  }
}
