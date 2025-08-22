import 'dart:io' show Platform;
import 'dart:convert';
import 'package:device_info_plus/device_info_plus.dart';
import 'package:crypto/crypto.dart';
import 'package:flutter/foundation.dart' show kIsWeb;

/// Basic cross-platform device info + fingerprint helpers.
class DeviceHelper {
  static final DeviceInfoPlugin _plugin = DeviceInfoPlugin();

  /// Returns a simple platform string: android/ios/web/windows/macos/linux
  static String platformType({String? override}) {
    if (override != null && override.isNotEmpty) return override;
    if (kIsWeb) return 'web';
    if (Platform.isAndroid) return 'android';
    if (Platform.isIOS) return 'ios';
    if (Platform.isWindows) return 'windows';
    if (Platform.isMacOS) return 'macos';
    if (Platform.isLinux) return 'linux';
    return 'unknown';
  }

  /// Collects a subset of stable device fields per platform for fingerprinting and display.
  static Future<Map<String, dynamic>> getDeviceInfo() async {
    try {
      if (kIsWeb) {
        final info = await _plugin.webBrowserInfo;
        return {
          'platform': 'web',
          'userAgent': info.userAgent,
          'vendor': info.vendor,
          'browserName': info.browserName.name,
          'appCodeName': info.appCodeName,
          'appName': info.appName,
          'deviceMemory': info.deviceMemory,
          'hardwareConcurrency': info.hardwareConcurrency,
        };
      }
      if (Platform.isAndroid) {
        final info = await _plugin.androidInfo;
        return {
          'platform': 'android',
          'brand': info.brand,
          'model': info.model,
          'manufacturer': info.manufacturer,
          'device': info.device,
          'id': info.id,
          'product': info.product,
          'hardware': info.hardware,
          'versionSdkInt': info.version.sdkInt,
        };
      }
      if (Platform.isIOS) {
        final info = await _plugin.iosInfo;
        return {
          'platform': 'ios',
          'name': info.name,
          'systemName': info.systemName,
          'systemVersion': info.systemVersion,
          'model': info.model,
          'localizedModel': info.localizedModel,
          'utsnameMachine': info.utsname.machine,
        };
      }
      if (Platform.isWindows) {
        final info = await _plugin.windowsInfo;
        return {
          'platform': 'windows',
          'computerName': info.computerName,
          'numberOfCores': info.numberOfCores,
          'systemMemoryInMegabytes': info.systemMemoryInMegabytes,
          'userName': info.userName,
        };
      }
      if (Platform.isMacOS) {
        final info = await _plugin.macOsInfo;
        return {
          'platform': 'macos',
          'computerName': info.computerName,
          'model': info.model,
          'arch': info.arch,
          'osRelease': info.osRelease,
        };
      }
      if (Platform.isLinux) {
        final info = await _plugin.linuxInfo;
        return {
          'platform': 'linux',
          'name': info.name,
          'version': info.version,
          'id': info.id,
          'prettyName': info.prettyName,
          'machineId': info.machineId,
        };
      }
    } catch (_) {
      // fallthrough to unknown
    }
    return {'platform': 'unknown'};
  }

  /// Computes a SHA256 hex fingerprint from selected device info fields.
  static String computeFingerprint(Map<String, dynamic> info) {
    // Sort keys for stability and concatenate key=value pairs
    final keys = info.keys.toList()..sort();
    final buffer = StringBuffer();
    for (final k in keys) {
      final v = info[k];
      buffer.write(k);
      buffer.write('=');
      buffer.write(v?.toString() ?? '');
      buffer.write(';');
    }
    final bytes = utf8.encode(buffer.toString());
    return sha256.convert(bytes).toString();
  }
}

/// Helper to build standard payload fields for login requests.
/// Returns: {
///   'device_id': `fingerprint`,
///   'device_name': `model/computerName/browserName`,
///   'device_type': `platformType`
/// }
Future<Map<String, String>> getCurrentDevicePayload() async {
  final info = await DeviceHelper.getDeviceInfo();
  final fp = DeviceHelper.computeFingerprint(info);
  final type = DeviceHelper.platformType(override: info['platform']?.toString());

  // Prefer a human-readable name depending on platform
  String? name = info['model']?.toString();
  name ??= info['computerName']?.toString();
  name ??= info['browserName']?.toString();
  name ??= info['name']?.toString();
  name ??= type;

  return {
    'device_id': fp,
    'device_name': name,
    'device_type': type,
  };
}

// Web detection handled by Flutter's kIsWeb above.
