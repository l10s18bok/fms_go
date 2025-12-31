import 'dart:convert';
import 'dart:io';
import 'package:path/path.dart' as path;
import '../models/models.dart';

/// JSON 파일 기반 저장소 서비스
class StorageService {
  static const String configDirName = 'config3';
  static const String templatesFileName = 'templates.json';
  static const String firewallsFileName = 'firewalls.json';
  static const String historyFileName = 'history.json';
  static const String configFileName = 'config.json';

  String? _configDir;
  int _nextFirewallIndex = 1;
  int _nextHistoryId = 1;

  /// 설정 디렉토리 경로 반환 (실행 파일 위치 기준)
  Future<String> getConfigDir() async {
    if (_configDir != null) return _configDir!;

    // 실행 파일 경로 기준으로 설정 디렉토리 설정
    final execPath = Platform.resolvedExecutable;
    final execDir = path.dirname(execPath);
    _configDir = path.join(execDir, configDirName);

    final dir = Directory(_configDir!);
    if (!await dir.exists()) {
      await dir.create(recursive: true);
    }

    return _configDir!;
  }

  /// 파일 경로 생성
  Future<String> _getFilePath(String fileName) async {
    final configDir = await getConfigDir();
    return path.join(configDir, fileName);
  }

  /// JSON 파일 읽기
  Future<dynamic> _readJsonFile(String fileName) async {
    try {
      final filePath = await _getFilePath(fileName);
      final file = File(filePath);
      if (!await file.exists()) {
        return null;
      }
      final contents = await file.readAsString();
      if (contents.isEmpty) return null;
      return jsonDecode(contents);
    } catch (e) {
      return null;
    }
  }

  /// JSON 파일 쓰기
  Future<void> _writeJsonFile(String fileName, dynamic data) async {
    final filePath = await _getFilePath(fileName);
    final file = File(filePath);
    await file.writeAsString(jsonEncode(data));
  }

  // ==================== 템플릿 관련 ====================

  /// 모든 템플릿 조회
  Future<List<Template>> getAllTemplates() async {
    final data = await _readJsonFile(templatesFileName);
    if (data == null) return [];
    return (data as List)
        .map((e) => Template.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  /// 단일 템플릿 조회
  Future<Template?> getTemplate(String version) async {
    final templates = await getAllTemplates();
    try {
      return templates.firstWhere((t) => t.version == version);
    } catch (e) {
      return null;
    }
  }

  /// 템플릿 저장 (추가/수정)
  Future<void> saveTemplate(String version, String contents) async {
    final templates = await getAllTemplates();
    final existingIndex = templates.indexWhere((t) => t.version == version);

    if (existingIndex >= 0) {
      templates[existingIndex] =
          Template(version: version, contents: contents);
    } else {
      templates.add(Template(version: version, contents: contents));
    }

    await _writeJsonFile(
        templatesFileName, templates.map((t) => t.toJson()).toList());
  }

  /// 템플릿 삭제
  Future<void> deleteTemplate(String version) async {
    final templates = await getAllTemplates();
    templates.removeWhere((t) => t.version == version);
    await _writeJsonFile(
        templatesFileName, templates.map((t) => t.toJson()).toList());
  }

  // ==================== 방화벽 장비 관련 ====================

  /// 모든 장비 조회
  Future<List<Firewall>> getAllFirewalls() async {
    final data = await _readJsonFile(firewallsFileName);
    if (data == null) return [];
    final firewalls = (data as List)
        .map((e) => Firewall.fromJson(e as Map<String, dynamic>))
        .toList();

    // 다음 인덱스 업데이트
    if (firewalls.isNotEmpty) {
      _nextFirewallIndex =
          firewalls.map((f) => f.index).reduce((a, b) => a > b ? a : b) + 1;
    }

    return firewalls;
  }

  /// 단일 장비 조회
  Future<Firewall?> getFirewall(int index) async {
    final firewalls = await getAllFirewalls();
    try {
      return firewalls.firstWhere((f) => f.index == index);
    } catch (e) {
      return null;
    }
  }

  /// 장비 저장 (추가/수정)
  Future<Firewall> saveFirewall(Firewall firewall) async {
    final firewalls = await getAllFirewalls();

    Firewall savedFirewall;
    if (firewall.index <= 0) {
      // 새 장비 추가
      savedFirewall = Firewall(
        index: _nextFirewallIndex++,
        deviceName: firewall.deviceName,
        serverStatus: firewall.serverStatus,
        deployStatus: firewall.deployStatus,
        version: firewall.version,
        deployResult: firewall.deployResult,
      );
      firewalls.add(savedFirewall);
    } else {
      // 기존 장비 수정
      final existingIndex =
          firewalls.indexWhere((f) => f.index == firewall.index);
      if (existingIndex >= 0) {
        firewalls[existingIndex] = firewall;
        savedFirewall = firewall;
      } else {
        firewalls.add(firewall);
        savedFirewall = firewall;
      }
    }

    await _writeJsonFile(
        firewallsFileName, firewalls.map((f) => f.toJson()).toList());
    return savedFirewall;
  }

  /// 장비 삭제
  Future<void> deleteFirewall(int index) async {
    final firewalls = await getAllFirewalls();
    firewalls.removeWhere((f) => f.index == index);
    await _writeJsonFile(
        firewallsFileName, firewalls.map((f) => f.toJson()).toList());
  }

  /// 장비 상태 업데이트
  Future<void> updateFirewallStatus(
    int index, {
    String? serverStatus,
    String? deployStatus,
    String? version,
    DeployResult? deployResult,
  }) async {
    final firewalls = await getAllFirewalls();
    final existingIndex = firewalls.indexWhere((f) => f.index == index);
    if (existingIndex >= 0) {
      final existing = firewalls[existingIndex];
      firewalls[existingIndex] = existing.copyWith(
        serverStatus: serverStatus ?? existing.serverStatus,
        deployStatus: deployStatus ?? existing.deployStatus,
        version: version ?? existing.version,
        deployResult: deployResult ?? existing.deployResult,
      );
      await _writeJsonFile(
          firewallsFileName, firewalls.map((f) => f.toJson()).toList());
    }
  }

  // ==================== 배포 이력 관련 ====================

  /// 모든 이력 조회
  Future<List<DeployHistory>> getAllHistory() async {
    final data = await _readJsonFile(historyFileName);
    if (data == null) return [];
    final history = (data as List)
        .map((e) => DeployHistory.fromJson(e as Map<String, dynamic>))
        .toList();

    // 다음 ID 업데이트
    if (history.isNotEmpty) {
      _nextHistoryId =
          history.map((h) => h.id).reduce((a, b) => a > b ? a : b) + 1;
    }

    // 최신순 정렬
    history.sort((a, b) => b.timestamp.compareTo(a.timestamp));
    return history;
  }

  /// 이력 저장
  Future<DeployHistory> saveHistory(DeployHistory historyItem) async {
    final history = await getAllHistory();

    DeployHistory savedHistory;
    if (historyItem.id <= 0) {
      savedHistory = DeployHistory(
        id: _nextHistoryId++,
        timestamp: historyItem.timestamp,
        deviceIp: historyItem.deviceIp,
        templateVersion: historyItem.templateVersion,
        status: historyItem.status,
        results: historyItem.results,
      );
      history.add(savedHistory);
    } else {
      savedHistory = historyItem;
      final existingIndex = history.indexWhere((h) => h.id == historyItem.id);
      if (existingIndex >= 0) {
        history[existingIndex] = historyItem;
      } else {
        history.add(historyItem);
      }
    }

    await _writeJsonFile(
        historyFileName, history.map((h) => h.toJson()).toList());
    return savedHistory;
  }

  /// 이력 삭제
  Future<void> deleteHistory(int id) async {
    final history = await getAllHistory();
    history.removeWhere((h) => h.id == id);
    await _writeJsonFile(
        historyFileName, history.map((h) => h.toJson()).toList());
  }

  /// 모든 이력 삭제
  Future<void> deleteAllHistory() async {
    await _writeJsonFile(historyFileName, []);
    _nextHistoryId = 1;
  }

  // ==================== 설정 관련 ====================

  /// 설정 조회
  Future<AppConfig> getConfig() async {
    final data = await _readJsonFile(configFileName);
    if (data == null) return AppConfig();
    return AppConfig.fromJson(data as Map<String, dynamic>);
  }

  /// 설정 저장
  Future<void> saveConfig(AppConfig config) async {
    await _writeJsonFile(configFileName, config.toJson());
  }

  // ==================== Import/Export ====================

  /// 모든 데이터 초기화
  Future<void> resetAll() async {
    await _writeJsonFile(templatesFileName, []);
    await _writeJsonFile(firewallsFileName, []);
    await _writeJsonFile(historyFileName, []);
    _nextFirewallIndex = 1;
    _nextHistoryId = 1;
  }

  /// 데이터 내보내기 (전체)
  Future<Map<String, dynamic>> exportAllData() async {
    final templates = await getAllTemplates();
    final firewalls = await getAllFirewalls();
    final history = await getAllHistory();
    final config = await getConfig();

    return {
      'templates': templates.map((t) => t.toJson()).toList(),
      'firewalls': firewalls.map((f) => f.toJson()).toList(),
      'history': history.map((h) => h.toJson()).toList(),
      'config': config.toJson(),
    };
  }

  /// 데이터 가져오기 (전체)
  Future<void> importAllData(Map<String, dynamic> data) async {
    if (data['templates'] != null) {
      await _writeJsonFile(templatesFileName, data['templates']);
    }
    if (data['firewalls'] != null) {
      await _writeJsonFile(firewallsFileName, data['firewalls']);
    }
    if (data['history'] != null) {
      await _writeJsonFile(historyFileName, data['history']);
    }
    if (data['config'] != null) {
      await _writeJsonFile(configFileName, data['config']);
    }
  }
}
