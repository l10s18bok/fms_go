import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/models.dart';
import 'storage_service.dart';

/// 배포 서비스 - 방화벽 장비와 HTTP 통신
class DeployService {
  final StorageService _storage;
  late AppConfig _config;

  DeployService(this._storage);

  /// 설정 로드
  Future<void> loadConfig() async {
    _config = await _storage.getConfig();
  }

  /// 서버 상태 확인
  Future<String> checkServerStatus(Firewall firewall) async {
    try {
      await loadConfig();

      final url = _config.connectionMode == 'agent'
          ? '${_config.agentServerURL}/agent/req-respCheck'
          : 'http://${firewall.deviceName}/respCheck';

      final response = await http
          .get(Uri.parse(url))
          .timeout(Duration(seconds: _config.timeoutSeconds));

      if (response.statusCode == 200) {
        return 'running';
      } else {
        return 'error';
      }
    } on TimeoutException {
      return 'stop';
    } catch (e) {
      return 'error';
    }
  }

  /// 선택된 장비들 상태 확인 (병렬)
  Future<Map<int, String>> checkSelectedServerStatus(
      List<Firewall> firewalls) async {
    await loadConfig();

    final results = <int, String>{};
    final futures = <Future<void>>[];

    for (final fw in firewalls) {
      futures.add(() async {
        final status = await checkServerStatus(fw);
        results[fw.index] = status;
        await _storage.updateFirewallStatus(fw.index, serverStatus: status);
      }());
    }

    await Future.wait(futures);
    return results;
  }

  /// 템플릿 배포
  Future<DeployResult> deploy(Firewall firewall, Template template) async {
    try {
      await loadConfig();

      final url = _config.connectionMode == 'agent'
          ? '${_config.agentServerURL}/agent/req-deploy'
          : 'http://${firewall.deviceName}/deploy';

      // 요청 본문 구성
      final body = jsonEncode({
        'deviceIp': firewall.deviceName,
        'templateVersion': template.version,
        'contents': template.contents,
      });

      final response = await http
          .post(
            Uri.parse(url),
            headers: {'Content-Type': 'application/json'},
            body: body,
          )
          .timeout(Duration(seconds: _config.timeoutSeconds));

      if (response.statusCode == 200) {
        final responseData = jsonDecode(response.body);
        final results = <RuleResult>[];

        // 응답 파싱
        if (responseData['results'] != null) {
          for (final r in responseData['results'] as List) {
            results.add(RuleResult(
              rule: r['rule'] ?? '',
              text: r['text'] ?? '',
              status: r['status'] ?? '',
              reason: r['reason'] ?? '',
            ));
          }
        }

        final status = responseData['status'] ?? 'success';
        final deployResult = DeployResult(
          ip: firewall.deviceName,
          status: status,
          results: results,
        );

        // 장비 상태 업데이트
        await _storage.updateFirewallStatus(
          firewall.index,
          deployStatus: status,
          version: template.version,
          deployResult: deployResult,
        );

        // 이력 저장
        await _storage.saveHistory(DeployHistory(
          id: 0,
          timestamp: DateTime.now(),
          deviceIp: firewall.deviceName,
          templateVersion: template.version,
          status: status,
          results: results,
        ));

        return deployResult;
      } else {
        throw Exception('배포 실패: HTTP ${response.statusCode}');
      }
    } on TimeoutException {
      final deployResult = DeployResult(
        ip: firewall.deviceName,
        status: 'error',
        results: [
          RuleResult(
            rule: '',
            text: '',
            status: 'error',
            reason: '타임아웃',
          )
        ],
      );

      await _storage.updateFirewallStatus(
        firewall.index,
        deployStatus: 'error',
        deployResult: deployResult,
      );

      await _storage.saveHistory(DeployHistory(
        id: 0,
        timestamp: DateTime.now(),
        deviceIp: firewall.deviceName,
        templateVersion: template.version,
        status: 'error',
        results: deployResult.results,
      ));

      return deployResult;
    } catch (e) {
      final deployResult = DeployResult(
        ip: firewall.deviceName,
        status: 'fail',
        results: [
          RuleResult(
            rule: '',
            text: '',
            status: 'error',
            reason: e.toString(),
          )
        ],
      );

      await _storage.updateFirewallStatus(
        firewall.index,
        deployStatus: 'fail',
        deployResult: deployResult,
      );

      await _storage.saveHistory(DeployHistory(
        id: 0,
        timestamp: DateTime.now(),
        deviceIp: firewall.deviceName,
        templateVersion: template.version,
        status: 'fail',
        results: deployResult.results,
      ));

      return deployResult;
    }
  }
}
