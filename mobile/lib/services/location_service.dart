import 'package:flutter/services.dart' show rootBundle;
import 'package:csv/csv.dart';
import '../models/location.dart';

class LocationService {
  static LocationService? _instance;
  factory LocationService() => _instance ??= LocationService._();
  LocationService._();

  List<Province>? _provinces;
  final Map<String, List<Ward>> _wardsByProvince = {};

  Future<void> _ensureLoaded() async {
    if (_provinces != null) return;

    final csv = await rootBundle.loadString('assets/vietnam_divisions.csv');
    final rows = const CsvToListConverter().convert(csv, eol: '\n');

    final provincesMap = <String, String>{};
    final wards = <Ward>[];

    for (int i = 3; i < rows.length; i++) {
      final row = rows[i];
      if (row.length < 8) continue;

      final provinceCode = row[2].toString().trim();
      final provinceName = row[3].toString().trim();
      final wardCode = row[6].toString().trim();
      final wardName = row[7].toString().trim();

      if (provinceCode.isEmpty || provinceName.isEmpty) continue;
      if (wardCode.isEmpty || wardName.isEmpty) continue;

      provincesMap[provinceCode] = provinceName;
      wards.add(Ward(code: wardCode, name: wardName, provinceCode: provinceCode));
    }

    _provinces = provincesMap.entries
        .map((e) => Province(code: e.key, name: e.value))
        .toList()
      ..sort((a, b) => a.name.compareTo(b.name));

    for (final ward in wards) {
      _wardsByProvince.putIfAbsent(ward.provinceCode, () => []).add(ward);
    }
  }

  Future<List<Province>> getProvinces() async {
    await _ensureLoaded();
    return _provinces!;
  }

  Future<List<Ward>> getWards(String provinceCode) async {
    await _ensureLoaded();
    return _wardsByProvince[provinceCode] ?? [];
  }

  Future<List<Ward>> searchWards(String provinceCode, String query) async {
    final wards = await getWards(provinceCode);
    if (query.isEmpty) return wards;
    final q = _removeDiacritics(query.toLowerCase());
    return wards.where((w) => _removeDiacritics(w.name.toLowerCase()).contains(q)).toList();
  }

  static String _removeDiacritics(String str) {
    const withDiacritics =    'àáảãạăắằẳẵặâấầẩẫậèéẻẽẹêếềểễệìíỉĩịòóỏõọôốồổỗộơớờởỡợùúủũụưứừửữựỳýỷỹỵđ';
    const withoutDiacritics = 'aaaaaaaaaaaaaaaaaeeeeeeeeeeeiiiiiooooooooooooooooouuuuuuuuuuuyyyyyd';
    var result = str;
    for (int i = 0; i < withDiacritics.length; i++) {
      result = result.replaceAll(withDiacritics[i], withoutDiacritics[i]);
    }
    return result;
  }
}
